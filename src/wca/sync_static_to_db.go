package wca

import (
	"fmt"
	"runtime"
	"sort"
	"time"

	"log"

	"github.com/guojia99/cubing-pro/src/internel/database/model/wca/utils"
	"github.com/guojia99/cubing-pro/src/wca/types"
)

type personBest struct {
	PersonId  string
	Country   string
	Continent string
	Single    int
	Avg       int
	Rank      int
}

func (s *syncer) selectComps() (comps []types.Competition, err error) {
	// 只选择需要的字段
	err = s.db.Select("id", "country_id", "year", "month", "day", "end_year", "end_month", "end_day").
		Find(&comps).Error
	if err != nil {
		return nil, err
	}

	// 按结束日期升序排序：先年，再月，再日
	sort.Slice(comps, func(i, j int) bool {
		if comps[i].EndYear != comps[j].EndYear {
			return comps[i].EndYear < comps[j].EndYear
		}
		if comps[i].EndMonth != comps[j].EndMonth {
			return comps[i].EndMonth < comps[j].EndMonth
		}
		return comps[i].EndDay < comps[j].EndDay
	})

	return comps, nil
}

func (s *syncer) countryMap() map[string]types.Country {
	var country []types.Country
	s.db.Find(&country)

	var out = make(map[string]types.Country)
	for _, v := range country {
		out[v.ID] = v
	}
	return out
}

const startYear = 2003

func getEndCompsTimer(t time.Time) time.Time {
	year := t.Year()
	startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, t.Location())
	endOfYear := time.Date(year, 12, 31, 0, 0, 0, 0, t.Location())

	// 计算 t 距离年初的天数（0 表示 1月1日）
	offsetDays := int(t.Sub(startOfYear).Hours() / 24)

	// 所在周索引（从0开始）
	weekIndex := offsetDays / 7

	// 该周最后一天 = 年初 + (weekIndex + 1) * 7 - 1
	candidate := startOfYear.AddDate(0, 0, (weekIndex+1)*7-1)

	// 不能超过12月31日
	if candidate.After(endOfYear) {
		return endOfYear
	}
	return candidate
}

const setStaticPersonRankWithTimerIndex = `
-- 必需的基础索引
CREATE INDEX idx_wca_id ON static_with_timer_ranks (wca_id);

-- 查询最终排名数据所需索引
CREATE INDEX idx_event_year_month_country ON static_with_timer_ranks (event_id, year, month, country);

-- 排名查询所需索引 (平均值排名)
CREATE INDEX idx_avg_rankings ON static_with_timer_ranks (event_id, year, month, avg_country_rank, avg_world_rank);

-- 排名查询所需索引 (单次排名)
CREATE INDEX idx_single_rankings ON static_with_timer_ranks (event_id, year, month, single_country_rank, single_world_rank);

-- 组合索引用于完整查询场景
CREATE INDEX idx_full_query ON static_with_timer_ranks (event_id, year, month, country, single_world_rank, avg_world_rank);
`

func (s *syncer) getResultMapWithEvent(eventId string) map[string][]types.Result {
	var results []types.Result

	// 只查询需要的字段
	s.db.Select("competition_id", "event_id", "best", "average", "person_id", "person_country_id").
		Where("event_id = ?", eventId).
		Find(&results)

	out := make(map[string][]types.Result)
	for _, v := range results {
		if _, ok := out[v.CompetitionID]; !ok {
			out[v.CompetitionID] = make([]types.Result, 0, 1) // 预分配小容量
		}
		out[v.CompetitionID] = append(out[v.CompetitionID], v)
	}

	results = nil // GC
	return out
}

func (s *syncer) getEvents() []types.Event {
	var events []types.Event
	s.db.Find(&events)
	return events
}

const monthSp = 1

type curPersonValue struct {
	WcaId   string
	Single  int
	Average int

	Country   string
	Continent string

	Rank int
}

func (s *syncer) updateStaticCurAllPersonValue(curAllPersonValue map[string]*curPersonValue, curTimeResults []types.Result, countryMap map[string]types.Country) {
	for _, res := range curTimeResults {
		if res.Best > 0 {
			if _, ok := curAllPersonValue[res.PersonID]; !ok {
				curAllPersonValue[res.PersonID] = &curPersonValue{
					WcaId:     res.PersonID,
					Single:    -1,
					Average:   -1,
					Country:   res.PersonCountryID,
					Continent: countryMap[res.PersonCountryID].ContinentID,
				}
			}
			if curAllPersonValue[res.PersonID].Single < 0 {
				curAllPersonValue[res.PersonID].Single = res.Best
			} else if utils.IsBestResult(res.EventID, res.Best, curAllPersonValue[res.PersonID].Single) {
				curAllPersonValue[res.PersonID].Single = res.Best
			}
		}
		if res.Average > 0 {
			if curAllPersonValue[res.PersonID].Average < 0 {
				curAllPersonValue[res.PersonID].Average = res.Average
			} else if utils.IsBestResult(res.EventID, res.Average, curAllPersonValue[res.PersonID].Average) {
				curAllPersonValue[res.PersonID].Average = res.Average
			}
		}
	}
}
func sortEventRanks(eventId string, avg bool, in []curPersonValue) {
	if len(in) == 0 {
		return
	}

	// 排序
	if avg {
		sort.Slice(in, func(i, j int) bool { return utils.IsBestResult(eventId, in[i].Average, in[j].Average) })
	} else {
		sort.Slice(in, func(i, j int) bool { return utils.IsBestResult(eventId, in[i].Single, in[j].Single) })
	}
	// 分配排名
	in[0].Rank = 1
	for i := 1; i < len(in); i++ {
		if avg {
			// 如果当前成绩和前一名相同，则并列
			if in[i].Average == in[i-1].Average {
				in[i].Rank = in[i-1].Rank
			} else {
				// 否则，排名 = 当前位置 + 1（因为 i 从 0 开始）
				in[i].Rank = i + 1
			}
		} else {
			// 如果当前成绩和前一名相同，则并列
			if in[i].Single == in[i-1].Single {
				in[i].Rank = in[i-1].Rank
			} else {
				// 否则，排名 = 当前位置 + 1（因为 i 从 0 开始）
				in[i].Rank = i + 1
			}
		}
	}
}

type rankingState struct {
	rank      int // 当前应分配的排名
	lastValue int // 上一个有效成绩
	prevRank  int // 上一次分配的 rank（用于并列）
	count     int // 已处理人数（用于计算下一个 rank）
}

func (s *syncer) getCurPersonsRankTimerSnapshots(
	eventID string,
	curTIme time.Time,
	curAllPersonValue map[string]*curPersonValue,
) []types.StaticWithTimerRank {
	var singleRanks = make([]curPersonValue, 0, len(curAllPersonValue))
	var avgRanks = make([]curPersonValue, 0, len(curAllPersonValue))
	var thisSnapshotsMap = make(map[string]types.StaticWithTimerRank, len(curAllPersonValue))

	for _, r := range curAllPersonValue {
		thisSnapshotsMap[r.WcaId] = types.StaticWithTimerRank{
			WcaID:   r.WcaId,
			EventID: eventID,
			Year:    curTIme.Year(),
			Month:   int(curTIme.Month()),
			Week:    (curTIme.Day() / 7) + 1,
			Single:  r.Single,
			Average: r.Average,
			Country: r.Country,
		}

		if r.Single > 0 {
			singleRanks = append(singleRanks, *r)
		}
		if r.Average > 0 {
			avgRanks = append(avgRanks, *r)
		}
	}

	sortEventRanks(eventID, false, singleRanks)
	sortEventRanks(eventID, true, avgRanks)

	countryState := make(map[string]rankingState)
	continentState := make(map[string]rankingState)
	// 单次数据
	for _, r := range singleRanks {
		p := thisSnapshotsMap[r.WcaId]
		p.Single = r.Single
		p.SingleWorldRank = r.Rank

		// NR
		state := countryState[r.Country]
		if state.count == 0 {
			// 第一个该国选手
			state.rank = 1
			state.lastValue = r.Single
		} else if r.Single == state.lastValue {
			// 并列
			state.rank = state.prevRank // 保持上次的 rank
		} else {
			// 新排名 = 已处理人数 + 1
			state.rank = state.count + 1
			state.lastValue = r.Single
		}
		state.prevRank = state.rank
		state.count++
		countryState[r.Country] = state
		p.SingleCountryRank = state.rank

		// AsR
		cState := continentState[r.Continent]
		if cState.count == 0 {
			cState.rank = 1
			cState.lastValue = r.Single
		} else if r.Single == cState.lastValue {
			cState.rank = cState.prevRank
		} else {
			cState.rank = cState.count + 1
			cState.lastValue = r.Single
		}
		cState.prevRank = cState.rank
		cState.count++
		continentState[r.Continent] = cState
		p.SingleContinentRank = cState.rank

		thisSnapshotsMap[r.WcaId] = p
	}

	countryState = make(map[string]rankingState)
	continentState = make(map[string]rankingState)
	// 平均数据
	for _, r := range avgRanks {
		p := thisSnapshotsMap[r.WcaId]
		p.Average = r.Average
		p.AvgWorldRank = r.Rank

		// NR
		state := countryState[r.Country]
		if state.count == 0 {
			// 第一个该国选手
			state.rank = 1
			state.lastValue = r.Average
		} else if r.Average == state.lastValue {
			// 并列
			state.rank = state.prevRank // 保持上次的 rank
		} else {
			// 新排名 = 已处理人数 + 1
			state.rank = state.count + 1
			state.lastValue = r.Average
		}
		state.prevRank = state.rank
		state.count++
		countryState[r.Country] = state
		p.AvgCountryRank = state.rank

		// AsR
		cState := continentState[r.Continent]
		if cState.count == 0 {
			cState.rank = 1
			cState.lastValue = r.Average
		} else if r.Average == cState.lastValue {
			cState.rank = cState.prevRank
		} else {
			cState.rank = cState.count + 1
			cState.lastValue = r.Average
		}
		cState.prevRank = cState.rank
		cState.count++
		continentState[r.Continent] = cState
		p.AvgContinentRank = cState.rank

		thisSnapshotsMap[r.WcaId] = p
	}

	var out = make([]types.StaticWithTimerRank, 0, len(thisSnapshotsMap))
	for _, ps := range thisSnapshotsMap {
		out = append(out, ps)
	}
	return out
}

func (s *syncer) setStaticPersonRankWithTimersWithEvent(eventID string, comps []types.Competition, countryMap map[string]types.Country, avg bool) error {
	resultMap := s.getResultMapWithEvent(eventID)

	var curAllPersonValue = make(map[string]*curPersonValue)

	idx := 0
	totalComps := len(comps)

	maxDay := getEndCompsTimer(time.Now()) // or your custom end date
	maxYear := maxDay.Year()
	for year := startYear; year <= maxYear; year++ {
		yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		yearEnd := time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)
		if yearStart.After(maxDay) {
			break
		}

		for month := time.January; month <= time.December; month += monthSp {
			monthStart := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
			if monthStart.After(maxDay) {
				break
			}
			monthEnd := monthStart.AddDate(0, monthSp, -1)
			if monthEnd.After(yearEnd) {
				monthEnd = yearEnd
			}
			if monthEnd.After(maxDay) {
				monthEnd = maxDay
			}

			var thisMonthResults []types.Result
			compCount := 0
			for idx < totalComps {
				c := comps[idx]
				compDate := time.Date(int(c.EndYear), time.Month(c.EndMonth), int(c.EndDay), 0, 0, 0, 0, time.UTC)

				if compDate.After(monthEnd) {
					break
				}
				if !compDate.Before(monthStart) {
					if results, ok := resultMap[c.ID]; ok {
						thisMonthResults = append(thisMonthResults, results...)
						compCount++
					}
				}
				idx++
			}

			if len(thisMonthResults) == 0 {
				continue
			}

			log.Printf("[Event %s] sync with time %s-%s | %d", eventID, monthStart.Format("2006-01"), monthEnd.Format("2006-01"), len(thisMonthResults))
			// 更新快照
			s.updateStaticCurAllPersonValue(curAllPersonValue, thisMonthResults, countryMap)
			thisMonthResults = nil
			runtime.GC()

			// 获取当前单次排名
			snapshots := s.getCurPersonsRankTimerSnapshots(eventID, monthEnd, curAllPersonValue)
			// 写入数据库
			if err := s.db.CreateInBatches(snapshots, 4500).Error; err != nil {
				return err
			}
			snapshots = nil
			runtime.GC()

			if monthEnd.Equal(maxDay) || monthEnd.After(maxDay) {
				goto endLoop
			}
		}

	}

endLoop:
	curAllPersonValue = nil
	runtime.GC()
	return nil
}

func (s *syncer) setStaticPersonRankWithTimers() (err error) {
	startTime := time.Now()
	err = s.db.AutoMigrate(&types.StaticWithTimerRank{})
	if err != nil {
		return
	}

	s.db.Delete(&types.StaticWithTimerRank{}, "1 = 1")

	defer func() {
		if err != nil {
			s.db.Delete(&types.StaticWithTimerRank{}, "1 = 1")
		}
	}()
	comps, err := s.selectComps()
	if err != nil {
		return fmt.Errorf("failed to select comps and results: %w", err)
	}
	if len(comps) == 0 {
		return fmt.Errorf("no comps found")
	}

	countryMap := s.countryMap()
	events := s.getEvents()
	for _, event := range events {
		log.Printf("start sync with event %s", event.ID)
		err = s.setStaticPersonRankWithTimersWithEvent(event.ID, comps, countryMap, false)
		if err != nil {
			return err
		}
	}

	log.Println("Adding index for StaticPersonRankWithTimer...")
	if err = s.syncAddIndex(s.currentDB, setStaticPersonRankWithTimerIndex); err != nil {
		return err
	}
	log.Printf("Monthly rank snapshot generation completed in %v", time.Since(startTime))
	return
}

func (s *syncer) setFirstRankTimer() error {
	return nil
}
