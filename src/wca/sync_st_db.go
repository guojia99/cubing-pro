package wca

import (
	"fmt"
	"sort"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/wca/utils"
	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/log"
	"github.com/guojia99/cubing-pro/src/wca/types"
)

type eventRank struct {
	PersonId string `json:"person_id"` // wcaId
	Value    int    `json:"best"`
	Rank     int    `json:"rank"`
}

type personBest struct {
	PersonId  string               `json:"person_id"`
	Country   string               // 国家
	Continent string               // 洲
	Single    map[string]eventRank `json:"s"`
	Avg       map[string]eventRank `json:"a"`
}

func (s *syncer) selectCompsAndResultMap() (comps []types.Competition, resultMap map[string][]types.Result, err error) {
	var results []types.Result
	if err = s.db.Find(&results).Error; err != nil {
		return
	}
	if err = s.db.Find(&comps).Error; err != nil {
		return
	}

	// 排序
	sort.Slice(comps, func(i, j int) bool {
		if comps[i].EndYear != comps[j].EndYear {
			return comps[i].EndYear < comps[j].EndYear
		}
		if comps[i].EndMonth != comps[j].EndMonth {
			return comps[i].EndMonth < comps[j].EndMonth
		}
		return comps[i].EndDay < comps[j].EndDay
	})

	resultMap = make(map[string][]types.Result)
	for _, result := range results {
		if _, ok := resultMap[result.CompetitionID]; !ok {
			resultMap[result.CompetitionID] = make([]types.Result, 0)
		}
		resultMap[result.CompetitionID] = append(resultMap[result.CompetitionID], result)
	}
	results = nil // GC
	return
}

func (s *syncer) countryMap() map[string]types.Country {
	var country []types.Country
	s.db.Find(&country)

	var out = make(map[string]types.Country)
	for _, v := range country {
		out[v.Name] = v
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
CREATE INDEX idx_wca_id ON static_person_rank_with_timer (wca_id);
`

func (s *syncer) setStaticPersonRankWithTimer() error {
	if err := s.db.AutoMigrate(&types.StaticPersonRankWithTimer{}); err != nil {
		return err
	}

	// todo 结束时如果返回错误，需要删除
	var err error
	defer func() {
		if err != nil {
			s.db.Delete(&types.StaticPersonRankWithTimer{}, "1 = 1")
		}
	}()

	startTime := time.Now()
	comps, resultMap, err := s.selectCompsAndResultMap()
	if err != nil {
		return fmt.Errorf("failed to select comps and results: %w", err)
	}
	countryMap := s.countryMap()

	if len(comps) == 0 {
		return fmt.Errorf("no comps found")
	}

	var curAllPersonValue = make(map[string]*personBest)

	maxDay := getEndCompsTimer(time.Now()) // or your custom end date
	maxYear := maxDay.Year()

	idx := 0
	totalComps := len(comps)

	log.Infof("Starting monthly rank snapshot generation from year %d to %d (maxDay: %s), total competitions: %d",
		startYear, maxYear, maxDay.Format("2006-01-02"), totalComps)

	for year := startYear; year <= maxYear; year++ {
		yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		yearEnd := time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)

		if yearStart.After(maxDay) {
			break
		}

		for month := time.January; month <= time.December; month++ {
			monthStart := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
			if monthStart.After(maxDay) {
				break
			}

			monthEnd := monthStart.AddDate(0, 1, -1)
			if monthEnd.After(yearEnd) {
				monthEnd = yearEnd
			}
			if monthEnd.After(maxDay) {
				monthEnd = maxDay
			}

			// 收集本月比赛
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

			log.Debugf("Processing %s: found %d competitions with %d results",
				monthEnd.Format("2006-01"), compCount, len(thisMonthResults))

			// 生成快照
			snapshotStart := time.Now()
			snapshots := s.getStaticPersonRankWithTimerUpdateStaticPersonRankWithTimer(
				countryMap,
				monthEnd,
				curAllPersonValue,
				thisMonthResults,
			)
			snapshotDuration := time.Since(snapshotStart)

			snapshotCount := len(snapshots)
			if snapshotCount > 0 {
				saveStart := time.Now()
				if err = s.db.CreateInBatches(snapshots, 8000).Error; err != nil {
					log.Errorf("failed to save snapshots: %+v", err)
					return err
				}
				saveDuration := time.Since(saveStart)
				log.Infof("Saved %d rank snapshots for %s (gen: %v, save: %v)",
					snapshotCount, monthEnd.Format("2006-01"), snapshotDuration, saveDuration)
			} else {
				log.Infof("No rank snapshots generated for %s (possibly no active persons)", monthEnd.Format("2006-01"))
			}

			// 提前终止
			if monthEnd.Equal(maxDay) || monthEnd.After(maxDay) {
				goto endLoop
			}
		}
	}
endLoop:
	if err = s.syncAddIndex(s.currentDB, setStaticPersonRankWithTimerIndex); err != nil {
		return err
	}
	log.Infof("Monthly rank snapshot generation completed in %v", time.Since(startTime))
	return nil
}

func (s *syncer) getStaticPersonRankWithTimerUpdateCurAllPersonValue(
	countryMap map[string]types.Country,
	curAllPersonValue map[string]*personBest,
	curWeekResults []types.Result) {
	// 当前所有选手成绩
	for _, res := range curWeekResults {
		if _, ok := curAllPersonValue[res.PersonID]; !ok {
			curAllPersonValue[res.PersonID] = &personBest{
				PersonId:  res.PersonID,
				Country:   res.PersonCountryID,
				Continent: countryMap[res.PersonCountryID].ContinentID,
				Single:    make(map[string]eventRank),
				Avg:       make(map[string]eventRank),
			}
		}
		if res.Best > 0 {
			if _, ok := curAllPersonValue[res.PersonID].Single[res.EventID]; !ok {
				curAllPersonValue[res.PersonID].Single[res.EventID] = eventRank{
					PersonId: res.PersonID,
					Value:    res.Best,
				}
			} else if utils.IsBestResult(res.EventID, res.Best, curAllPersonValue[res.PersonID].Single[res.EventID].Value) {
				curAllPersonValue[res.PersonID].Single[res.EventID] = eventRank{
					PersonId: res.PersonID,
					Value:    res.Best,
				}
			}
		}
		if res.Average > 0 {
			if _, ok := curAllPersonValue[res.PersonID].Avg[res.EventID]; !ok {
				curAllPersonValue[res.PersonID].Avg[res.EventID] = eventRank{
					PersonId: res.PersonID,
					Value:    res.Average,
				}
			} else if utils.IsBestResult(res.EventID, res.Average, curAllPersonValue[res.PersonID].Avg[res.EventID].Value) {
				curAllPersonValue[res.PersonID].Avg[res.EventID] = eventRank{
					PersonId: res.PersonID,
					Value:    res.Average,
				}
			}
		}
	}
}

func sortEventRanks(in map[string][]eventRank) map[string][]eventRank {
	for event, list := range in {
		if len(list) == 0 {
			continue
		}

		// 排序
		sort.Slice(list, func(i, j int) bool {
			return utils.IsBestResult(event, list[i].Value, list[j].Value)
		})

		// 分配排名
		list[0].Rank = 1
		for i := 1; i < len(list); i++ {
			// 如果当前成绩和前一名相同，则并列
			if list[i].Value == list[i-1].Value {
				list[i].Rank = list[i-1].Rank
			} else {
				// 否则，排名 = 当前位置 + 1（因为 i 从 0 开始）
				list[i].Rank = i + 1
			}
		}

		in[event] = list
	}
	return in
}

type rankingState struct {
	rank      int // 当前应分配的排名
	lastValue int // 上一个有效成绩
	prevRank  int // 上一次分配的 rank（用于并列）
	count     int // 已处理人数（用于计算下一个 rank）
}

func (s *syncer) getStaticPersonRankWithTimerUpdateStaticPersonRankWithTimer(
	countryMap map[string]types.Country,
	curWeekTIme time.Time,
	curAllPersonValue map[string]*personBest,
	curWeekResults []types.Result,
) []types.StaticPersonRankWithTimer {
	if len(curWeekResults) == 0 {
		return nil
	}
	// 更新curAllPersonValue
	s.getStaticPersonRankWithTimerUpdateCurAllPersonValue(countryMap, curAllPersonValue, curWeekResults)

	// 生成所有的列表
	var singleResultWithEventRank = make(map[string][]eventRank)
	var avgResultWithEventRank = make(map[string][]eventRank)
	var outMap = make(map[string]types.StaticPersonRankWithTimer)
	for _, r := range curAllPersonValue {
		outMap[r.PersonId] = types.StaticPersonRankWithTimer{
			ID:    0,
			WcaID: r.PersonId,
			Year:  curWeekTIme.Year(),
			Month: int(curWeekTIme.Month()),
			//Week:      (curWeekTIme.Day() / 7) + 1, // day从0开始算的
			Country:   r.Country,
			Continent: r.Continent,
			Ranks: types.StaticPersonRankWithTimerRanks{
				Single: &types.StaticPersonRank{
					CountryRank:   make(map[string]int), // 不同项目
					WorldRank:     make(map[string]int),
					ContinentRank: make(map[string]int),
				},
				Avg: &types.StaticPersonRank{
					CountryRank:   make(map[string]int),
					WorldRank:     make(map[string]int),
					ContinentRank: make(map[string]int),
				},
			},
		}

		for e, v := range r.Single {
			if _, ok := singleResultWithEventRank[e]; !ok {
				singleResultWithEventRank[e] = make([]eventRank, 0)
			}
			singleResultWithEventRank[e] = append(singleResultWithEventRank[e], v)
		}
		for e, v := range r.Avg {
			if _, ok := avgResultWithEventRank[e]; !ok {
				avgResultWithEventRank[e] = make([]eventRank, 0)
			}
			avgResultWithEventRank[e] = append(avgResultWithEventRank[e], v)
		}
	}

	// 全局进行排序
	singleResultWithEventRank = sortEventRanks(singleResultWithEventRank)
	avgResultWithEventRank = sortEventRanks(avgResultWithEventRank)

	// 根据不同国家、洲， 以及全部（WR）来填充outMap 中不同人的快照排名
	for e, list := range singleResultWithEventRank {
		// 国家 -> 当前排名状态
		countryState := make(map[string]rankingState)
		continentState := make(map[string]rankingState)

		for _, r := range list {
			p := outMap[r.PersonId]

			// WR 已由 sortEventRanks 设置，直接使用
			p.Ranks.Single.WorldRank[e] = r.Rank

			// --- National Rank (NR) ---
			state := countryState[p.Country]
			if state.count == 0 {
				// 第一个该国选手
				state.rank = 1
				state.lastValue = r.Value
			} else if r.Value == state.lastValue {
				// 并列
				state.rank = state.prevRank // 保持上次的 rank
			} else {
				// 新排名 = 已处理人数 + 1
				state.rank = state.count + 1
				state.lastValue = r.Value
			}
			state.prevRank = state.rank
			state.count++
			countryState[p.Country] = state
			p.Ranks.Single.CountryRank[e] = state.rank

			// --- Continental Rank (CR) ---
			cState := continentState[p.Continent]
			if cState.count == 0 {
				cState.rank = 1
				cState.lastValue = r.Value
			} else if r.Value == cState.lastValue {
				cState.rank = cState.prevRank
			} else {
				cState.rank = cState.count + 1
				cState.lastValue = r.Value
			}
			cState.prevRank = cState.rank
			cState.count++
			continentState[p.Continent] = cState
			p.Ranks.Single.ContinentRank[e] = cState.rank
		}
	}

	// 处理 avg（完全对称）
	for e, list := range avgResultWithEventRank {
		countryState := make(map[string]rankingState)
		continentState := make(map[string]rankingState)

		for _, r := range list {
			p := outMap[r.PersonId]

			p.Ranks.Avg.WorldRank[e] = r.Rank

			// National Rank
			state := countryState[p.Country]
			if state.count == 0 {
				state.rank = 1
				state.lastValue = r.Value
			} else if r.Value == state.lastValue {
				state.rank = state.prevRank
			} else {
				state.rank = state.count + 1
				state.lastValue = r.Value
			}
			state.prevRank = state.rank
			state.count++
			countryState[p.Country] = state
			p.Ranks.Avg.CountryRank[e] = state.rank

			// Continental Rank
			cState := continentState[p.Continent]
			if cState.count == 0 {
				cState.rank = 1
				cState.lastValue = r.Value
			} else if r.Value == cState.lastValue {
				cState.rank = cState.prevRank
			} else {
				cState.rank = cState.count + 1
				cState.lastValue = r.Value
			}
			cState.prevRank = cState.rank
			cState.count++
			continentState[p.Continent] = cState
			p.Ranks.Avg.ContinentRank[e] = cState.rank
		}
	}

	var out []types.StaticPersonRankWithTimer
	for _, v := range outMap {
		out = append(out, v)
	}
	return out
}
