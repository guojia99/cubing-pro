package wca

import (
	"container/heap"
	"fmt"
	"log"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/wca/utils"
	"github.com/guojia99/cubing-pro/src/wca/types"
	utils_tool "github.com/guojia99/cubing-pro/src/wca/utils"
)

const wcaStartYear = 2003

func (s *syncer) getAllPersonYearMap() map[int]map[string]types.Person {
	var out = make(map[int]map[string]types.Person)

	getPersonIDYear := func(id string) (int, error) {
		if len(id) < 4 {
			return 0, fmt.Errorf("string length less than 4")
		}

		last4 := id[:4]
		n, err := strconv.Atoi(last4)
		if err != nil {
			return 0, err
		}
		if n <= wcaStartYear {
			return wcaStartYear, nil
		}
		return n, nil
	}

	var persons []types.Person
	if err := s.db.Where("sub_id = 1").Find(&persons).Error; err != nil {
		return nil
	}
	for _, person := range persons {
		n, err := getPersonIDYear(person.WcaID)
		if err != nil {
			fmt.Println(err)
			break
		}
		if _, ok := out[n]; !ok {
			out[n] = make(map[string]types.Person)
		}
		out[n][person.WcaID] = person
	}

	return out
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

func (s *syncer) getAttemptMap() map[int64][]int64 {
	var attemptMap = make(map[int64][]int64)

	var att []types.ResultAttempt
	s.db.Select("value", "result_id").Find(&att)

	for _, at := range att {
		if _, ok := attemptMap[at.ResultID]; !ok {
			attemptMap[at.ResultID] = make([]int64, 0)
		}
		attemptMap[at.ResultID] = append(attemptMap[at.ResultID], at.Value)
	}
	att = nil // GC
	runtime.GC()
	return attemptMap
}

func (s *syncer) setStaticSuccessRateResultWithEvent(eventID string) error {
	var results []types.Result
	s.db.Select("id", "person_id", "person_country_id", "person_name").Where("event_id = ?", eventID).Find(&results)

	attemptMap := make(map[int64][]int64)
	for i := 0; i < len(results); i += 10000 {
		end := i + 10000
		if end > len(results) {
			end = len(results)
		}

		batch := results[i:end]

		// 提取 ID 列表
		ids := make([]int64, len(batch))
		for j, r := range batch {
			ids[j] = r.ID
		}
		// 一次性查询该批次所有 attempts
		var attempts []types.ResultAttempt
		s.db.Select("value", "result_id").Where("result_id IN ?", ids).Find(&attempts)

		// 构建 map
		for _, at := range attempts {
			if _, ok := attemptMap[at.ResultID]; !ok {
				attemptMap[at.ResultID] = make([]int64, 0)
			}
			attemptMap[at.ResultID] = append(attemptMap[at.ResultID], at.Value)
		}
	}

	var personData = make(map[string]*types.StaticSuccessRateResult)
	for _, result := range results {
		if _, ok := personData[result.PersonID]; !ok {
			personData[result.PersonID] = &types.StaticSuccessRateResult{
				WcaID:      result.PersonID,
				WcaName:    result.PersonName,
				Country:    result.PersonCountryID,
				EventID:    eventID,
				Solved:     0,
				Attempted:  0,
				Percentage: 0,
			}
		}

		attempts, ok := attemptMap[result.ID]
		if !ok {
			continue
		}
		for _, attempt := range attempts {
			personData[result.PersonID].Attempted += 1
			if attempt > 0 {
				personData[result.PersonID].Solved += 1
			}
		}
		if personData[result.PersonID].Attempted >= 1 {
			personData[result.PersonID].Percentage = float64(personData[result.PersonID].Solved) / float64(personData[result.PersonID].Attempted)
		}
	}

	var saveData []types.StaticSuccessRateResult
	for _, result := range personData {
		saveData = append(saveData, *result)
	}

	log.Printf("save event %s (%d) count", eventID, len(saveData))
	if err := s.db.CreateInBatches(saveData, 5000).Error; err != nil {
		return err
	}
	return nil
}

const setSuccessRateResultWithEventIndex = `
-- 必需的基础索引
CREATE INDEX idx_event ON static_success_rate_results (event_id);
CREATE INDEX idx_event_country ON static_success_rate_results (event_id, country);
`

func (s *syncer) setStaticSuccessRateResult() (err error) {
	startTime := time.Now()
	err = s.db.AutoMigrate(&types.StaticSuccessRateResult{})
	if err != nil {
		return
	}
	s.db.Delete(&types.StaticSuccessRateResult{}, "1 = 1")
	defer func() {
		if err != nil {
			s.db.Delete(&types.StaticSuccessRateResult{}, "1 = 1")
		}
	}()
	log.Printf("get attempt start")
	events := []string{
		"333bf", "444bf", "555bf", "333fm", "clock",
	}
	for _, event := range events {
		log.Printf("start sync with event %s", event)
		err = s.setStaticSuccessRateResultWithEvent(event)
		if err != nil {
			return err
		}
	}
	log.Println("Adding index for StaticSuccessRateResult...")
	if err = s.syncAddIndex(s.currentDB, setSuccessRateResultWithEventIndex); err != nil {
		return err
	}
	log.Printf("Monthly rank snapshot generation completed in %v", time.Since(startTime))
	return nil
}

func (s *syncer) extendAllEventAvgPersonResults(ps types.AllEventAvgPersonResults, compMap map[string]types.Competition) types.AllEventAvgPersonResults {
	var person types.Person
	if err := s.db.Where("wca_id = ? ", ps.WcaID).First(&person).Error; err != nil {
		return ps
	}

	ps.Name = person.Name
	ps.Country = person.CountryID
	if ps.LackNum != 0 {
		return ps
	}

	var results []types.Result
	s.db.Where("person_id = ?", ps.WcaID).Find(&results)

	newPs := types.AllEventAvgPersonResults{
		WcaID:         ps.WcaID,
		Name:          person.Name,
		Country:       person.CountryID,
		DoneEventList: ps.DoneEventList,
		LackNum:       0,
		IsDone:        true,
	}

	var checkList []string

	var checkResultsMap = make(map[string]bool)
	var useCompMap = make(map[string]bool)
	for _, result := range results {
		if !slices.Contains(wcaEventsList, result.EventID) {
			continue
		}

		if _, ok := checkResultsMap[result.EventID]; !ok {
			switch result.EventID {
			case "333mbf":
				if result.Best > 0 {
					checkList = append(checkList, result.EventID)
					checkResultsMap[result.EventID] = true
				}
			default:
				if result.Average > 0 {
					checkList = append(checkList, result.EventID)
					checkResultsMap[result.EventID] = true
				}
			}
			useCompMap[result.CompetitionID] = true
		}

		if len(checkResultsMap) == 1 && newPs.StartTime == nil {
			cp := compMap[result.CompetitionID]
			ts := time.Date(int(cp.Year), time.Month(cp.Month), int(cp.Day), 0, 0, 0, 0, time.UTC)
			newPs.StartTime = &ts
		}

		if len(checkResultsMap) == MapEventNum {
			newPs.CompID = result.CompetitionID
			cp := compMap[result.CompetitionID]
			newPs.CompName = cp.Name
			ts := time.Date(int(cp.EndYear), time.Month(cp.EndMonth), int(cp.EndDay), 0, 0, 0, 0, time.UTC)
			newPs.EndTime = &ts
			newPs.UseCompNum = len(useCompMap)

			duration := newPs.EndTime.Sub(*newPs.StartTime)
			newPs.UseDate = int(duration/(24*time.Hour)) + 1

			break
		}
	}
	return newPs
}

func (s *syncer) getCompMap() map[string]types.Competition {
	var cps []types.Competition
	var out = make(map[string]types.Competition)

	s.db.Select("id", "name", "year", "month", "day", "end_year", "end_month", "end_day").Find(&cps)

	for _, cp := range cps {
		out[cp.ID] = cp
	}
	return out
}

func (s *syncer) setStaticAllEventAvg() (err error) {
	err = s.db.AutoMigrate(&types.AllEventAvgPersonResults{})
	if err != nil {
		return
	}
	s.db.Delete(&types.AllEventAvgPersonResults{}, "1 = 1")
	defer func() {
		if err != nil {
			s.db.Delete(&types.AllEventAvgPersonResults{}, "1 = 1")
		}
	}()

	var allPersonResultsMap = make(map[string]*types.AllEventAvgPersonResults)

	var RanksSingleWith333Mbfs []types.RanksSingle // 只有多盲需要单次
	var RanksAvgs []types.RanksAverage
	s.db.Where("event_id = ?", "333mbf").Find(&RanksSingleWith333Mbfs)
	s.db.Find(&RanksAvgs)

	lackNum := MapEventNum

	for _, single := range RanksSingleWith333Mbfs {
		if _, ok := allPersonResultsMap[single.PersonID]; !ok {
			allPersonResultsMap[single.PersonID] = &types.AllEventAvgPersonResults{
				WcaID:         single.PersonID,
				LackNum:       lackNum,
				DoneEventList: []string{},
				IsDone:        false,
			}
		}

		findP, _ := allPersonResultsMap[single.PersonID]
		if slices.Contains(findP.DoneEventList, "333mbf") {
			continue
		}
		findP.DoneEventList = append(findP.DoneEventList, "333mbf")
		findP.LackNum -= 1
		allPersonResultsMap[single.PersonID] = findP
	}

	for _, avg := range RanksAvgs {
		if !slices.Contains(wcaEventsList, avg.EventID) {
			continue
		}

		if _, ok := allPersonResultsMap[avg.PersonID]; !ok {
			allPersonResultsMap[avg.PersonID] = &types.AllEventAvgPersonResults{
				WcaID:         avg.PersonID,
				LackNum:       lackNum,
				DoneEventList: []string{},
				IsDone:        false,
			}
		}

		findP, _ := allPersonResultsMap[avg.PersonID]
		if slices.Contains(findP.DoneEventList, avg.EventID) {
			continue
		}
		findP.DoneEventList = append(findP.DoneEventList, avg.EventID)
		findP.LackNum -= 1
		allPersonResultsMap[avg.PersonID] = findP
	}

	compMap := s.getCompMap()

	var cutNumEventResults []types.AllEventAvgPersonResults
	for _, res := range allPersonResultsMap {
		if res.LackNum <= 4 {
			cutNumEventResults = append(cutNumEventResults, s.extendAllEventAvgPersonResults(*res, compMap))
		}
	}

	if err = s.db.CreateInBatches(cutNumEventResults, 5000).Error; err != nil {
		log.Printf("create in batch failed, err:%v", err)
	}
	return nil
}

const (
	WorldType     = "world"
	ContinentType = "continent"
	CountryType   = "country"
)

func (s *syncer) getChampionshipMap() (map[string][]types.Championship, map[string]types.Competition) {
	var championships []types.Championship
	s.db.Find(&championships)

	// 锦标赛
	var championshipMap = map[string][]types.Championship{
		WorldType:     make([]types.Championship, 0),
		ContinentType: make([]types.Championship, 0),
		CountryType:   make([]types.Championship, 0),
	}
	var compIds []string
	for _, championship := range championships {
		compIds = append(compIds, championship.CompetitionID)
		if championship.ChampionshipType == "world" {
			championshipMap[WorldType] = append(championshipMap[WorldType], championship)
			continue
		}

		if championship.ChampionshipType[0] == '_' {
			championshipMap[ContinentType] = append(championshipMap[ContinentType], championship)
			continue
		}

		//// cn 特殊处理, 重复了
		//if championship.CompetitionID == "AsianChampionship2016" && championship.ChampionshipType == "greater_china" {
		//	continue
		//}

		championshipMap[CountryType] = append(championshipMap[CountryType], championship)
	}

	// 比赛
	var compsMap = make(map[string]types.Competition)
	var comps []types.Competition
	s.db.Where("id in ?", compIds).Find(&comps)
	for _, comp := range comps {
		compsMap[comp.ID] = comp
	}

	for key := range championshipMap {

		list := championshipMap[key]

		sort.Slice(list, func(i, j int) bool {
			compA := compsMap[list[i].CompetitionID]
			compB := compsMap[list[j].CompetitionID]
			// 比较年
			if compA.EndYear != compB.EndYear {
				return compA.EndYear < compB.EndYear
			}
			// 比较月
			if compA.EndMonth != compB.EndMonth {
				return compA.EndMonth < compB.EndMonth
			}
			// 比较日
			if compA.EndDay != compB.EndDay {
				return compA.EndDay < compB.EndDay
			}
			return compA.ID < compB.ID
		})
		championshipMap[key] = list
	}

	return championshipMap, compsMap
}

func (s *syncer) getCompsResultAndCutPosTop3Map(compIDs ...string) map[string][]types.Result {
	var results []types.Result
	s.db.Where("competition_id in ?", compIDs).Find(&results)

	var out = make(map[string][]types.Result)

	for _, result := range results {
		if !(result.RoundTypeID == "f" || result.RoundTypeID == "c") {
			continue
		}
		if result.Pos >= 4 {
			continue
		}
		if result.Best <= 0 {
			continue
		}
		if _, ok := out[result.EventID]; !ok {
			out[result.EventID] = make([]types.Result, 0)
		}

		out[result.EventID] = append(out[result.EventID], result)
	}

	return out
}

func (s *syncer) getCountryContinentMap() map[string]string {
	var countrys []types.Country
	s.db.Find(&countrys)

	var out = make(map[string]string)
	for _, country := range countrys {
		out[country.ID] = country.ContinentID
	}
	return out
}

func (s *syncer) setStaticAllEventChampionshipsPodium() (err error) {

	err = s.db.AutoMigrate(&types.AllEventChampionshipsPodium{})
	if err != nil {
		return
	}
	s.db.Delete(&types.AllEventChampionshipsPodium{}, "1 = 1")
	defer func() {
		if err != nil {
			s.db.Delete(&types.AllEventChampionshipsPodium{}, "1 = 1")
		}
	}()

	// 1. 先找世锦赛选手, 再依据世锦赛前三找洲际赛, 然后是国家赛
	// 2. 统计出合适的选手再插该选手在该三场比赛的成绩
	// 3. 记录全部成绩并写入
	championshipMap, compMap := s.getChampionshipMap()

	var cache = make(map[string]map[string]types.AllEventChampionshipsPodium) // key1 event, key2 personID

	// world
	var worldCompsID []string
	for _, wc := range championshipMap[WorldType] {
		worldCompsID = append(worldCompsID, wc.CompetitionID)
	}
	for _, wcs := range s.getCompsResultAndCutPosTop3Map(worldCompsID...) {
		for _, res := range wcs {
			if _, ok := cache[res.EventID]; !ok {
				cache[res.EventID] = make(map[string]types.AllEventChampionshipsPodium)
			}

			if _, ok := cache[res.EventID][res.PersonID]; !ok {
				cache[res.EventID][res.PersonID] = types.AllEventChampionshipsPodium{
					WcaID:                    res.PersonID,
					WcaName:                  res.PersonName,
					Country:                  res.PersonCountryID,
					EventID:                  res.EventID,
					WorldChampionshipID:      res.CompetitionID,
					WorldChampionshipName:    compMap[res.CompetitionID].Name,
					WorldChampionshipRank:    int(res.Pos),
					WorldChampionshipBest:    res.Best,
					WorldChampionshipAverage: res.Average,
				}
			}

			cur := cache[res.EventID][res.PersonID]

			if cur.WorldChampionshipID != "" && int(res.Pos) > cur.WorldChampionshipRank {
				continue
			}

			cur.WorldChampionshipID = res.CompetitionID
			cur.WorldChampionshipName = compMap[res.CompetitionID].Name
			cur.WorldChampionshipRank = int(res.Pos)
			cur.WorldChampionshipBest = res.Best
			cur.WorldChampionshipAverage = res.Average
		}
	}

	countryContinentMap := s.getCountryContinentMap()
	compContinentMap := make(map[string]string)
	// 洲
	var continentCompsID []string
	for _, wc := range championshipMap[ContinentType] {
		continentCompsID = append(continentCompsID, wc.CompetitionID)
		compContinentMap[wc.CompetitionID] = countryContinentMap[compMap[wc.CompetitionID].CountryID]
	}

	for _, ccs := range s.getCompsResultAndCutPosTop3Map(continentCompsID...) {
		for _, res := range ccs {
			// 没有的项目直接剔除
			if _, ok := cache[res.EventID]; !ok {
				continue
			}

			// 没有的选手直接剔除
			if _, ok := cache[res.EventID][res.PersonID]; !ok {
				continue
			}

			cur := cache[res.EventID][res.PersonID]

			// 限制所在洲
			personContinent := countryContinentMap[res.PersonCountryID]
			compContinent := compContinentMap[res.CompetitionID]
			if personContinent != compContinent {
				continue
			}

			if cur.ContinentChampionshipID != "" && int(res.Pos) >= cur.ContinentChampionshipRank {
				continue
			}

			// 后续限制自己的洲
			cur.ContinentChampionshipID = res.CompetitionID
			cur.ContinentChampionshipName = compMap[res.CompetitionID].Name
			cur.ContinentChampionshipRank = int(res.Pos)
			cur.ContinentChampionshipBest = res.Best
			cur.ContinentChampionshipAverage = res.Average
			cache[res.EventID][res.PersonID] = cur
		}
	}

	// 国家级
	var countryCompsID []string
	for _, wc := range championshipMap[CountryType] {
		countryCompsID = append(countryCompsID, wc.CompetitionID)
	}
	for _, ncs := range s.getCompsResultAndCutPosTop3Map(countryCompsID...) {
		for _, res := range ncs {
			if _, ok := cache[res.EventID]; !ok {
				continue
			}
			if _, ok := cache[res.EventID][res.PersonID]; !ok {
				continue
			}
			cur := cache[res.EventID][res.PersonID]
			if cur.CountryChampionshipID != "" {
				continue
			}
			if compMap[res.CompetitionID].CountryID != res.PersonCountryID {
				continue
			}

			if cur.CountryChampionshipID != "" && int(res.Pos) >= cur.CountryChampionshipRank {
				continue
			}

			cur.CountryChampionshipID = res.CompetitionID
			cur.CountryChampionshipName = compMap[res.CompetitionID].Name
			cur.CountryChampionshipRank = int(res.Pos)
			cur.CountryChampionshipBest = res.Best
			cur.CountryChampionshipAverage = res.Average
			cache[res.EventID][res.PersonID] = cur
		}
	}

	var data []types.AllEventChampionshipsPodium
	for _, ev := range cache {
		for _, res := range ev {
			if res.CountryChampionshipID == "" || res.ContinentChampionshipID == "" {
				continue
			}
			data = append(data, res)
		}
	}
	for idx, val := range data {
		// 最佳成绩
		var bestRank types.RanksSingle
		var avgRank types.RanksAverage

		s.db.Where("event_id = ? and person_id = ?", val.EventID, val.WcaID).First(&bestRank)
		s.db.Where("event_id = ? and person_id = ?", val.EventID, val.WcaID).First(&avgRank)

		data[idx].Best = bestRank.Best
		data[idx].Average = avgRank.Best

		// 检查WR
		var results []types.Result
		s.db.Where("event_id = ?", val.EventID).Where("person_id = ?", val.WcaID).Find(&results)
		for _, result := range results {
			if result.RegionalSingleRecord == "WR" || result.RegionalAverageRecord == "WR" {
				data[idx].HasWR = true
				break
			}
		}
	}

	s.db.CreateInBatches(data, 1000)
	return nil
}

func (s *syncer) getAllRankWithEventPersonMap() (map[string]map[string]types.RanksSingle, map[string]map[string]types.RanksAverage) {
	var allSingleRanks []types.RanksSingle
	var allAvgRanks []types.RanksAverage
	s.db.Find(&allSingleRanks)
	s.db.Find(&allAvgRanks)

	var singleMap = make(map[string]map[string]types.RanksSingle)
	var avgMap = make(map[string]map[string]types.RanksAverage)

	for _, rank := range allSingleRanks {
		if _, ok := singleMap[rank.EventID]; !ok {
			singleMap[rank.EventID] = make(map[string]types.RanksSingle)
		}
		singleMap[rank.EventID][rank.PersonID] = rank
	}

	for _, rank := range allAvgRanks {
		if _, ok := avgMap[rank.EventID]; !ok {
			avgMap[rank.EventID] = make(map[string]types.RanksAverage)
		}
		avgMap[rank.EventID][rank.PersonID] = rank
	}
	return singleMap, avgMap
}

func (s *syncer) setStaticRankWithPersonStartYear() (err error) {
	err = s.db.AutoMigrate(&types.RankWithPersonCompStartYear{})
	if err != nil {
		return
	}
	s.db.Delete(&types.RankWithPersonCompStartYear{}, "1 = 1")
	defer func() {
		if err != nil {
			s.db.Delete(&types.RankWithPersonCompStartYear{}, "1 = 1")
		}
	}()

	// 1. 全部选手的数据，按年区分
	personMap := s.getAllPersonYearMap()

	// 3. 获取全部最佳成绩
	singleMap, avgMap := s.getAllRankWithEventPersonMap()

	//  循环年份

	for _, event := range wcaEventsList {
		// 单次
		if singleWithEventList, ok := singleMap[event]; ok {
			for year := wcaStartYear; year <= time.Now().Year(); year++ {
				yearPersons := personMap[year]
				// 单次
				var results []types.RankWithPersonCompStartYear
				for _, p := range yearPersons {
					if rank, hasRank := singleWithEventList[p.WcaID]; hasRank {
						results = append(results, types.RankWithPersonCompStartYear{
							PersonID:    p.WcaID,
							PersonName:  p.Name,
							CountryID:   p.CountryID,
							Year:        year,
							IsAvg:       false,
							EventID:     event,
							Best:        rank.Best,
							WorldRank:   rank.WorldRank,
							CountryRank: rank.CountryRank,
						})
					}
				}
				s.db.CreateInBatches(results, 5000)
			}
		}

		if event == "333mbf" {
			continue
		}

		if avgWithEventList, ok := avgMap[event]; ok {
			for year := wcaStartYear; year <= time.Now().Year(); year++ {
				yearPersons := personMap[year]
				var results []types.RankWithPersonCompStartYear
				for _, p := range yearPersons {
					if rank, hasRank := avgWithEventList[p.WcaID]; hasRank {
						results = append(results, types.RankWithPersonCompStartYear{
							PersonID:    p.WcaID,
							PersonName:  p.Name,
							CountryID:   p.CountryID,
							Year:        year,
							IsAvg:       true,
							EventID:     event,
							Best:        rank.Best,
							WorldRank:   rank.WorldRank,
							CountryRank: rank.CountryRank,
						})
					}
				}
				s.db.CreateInBatches(results, 5000)
			}
		}
	}
	return
}

type personRanks struct {
	personID        string
	single          map[string]int
	singleEventCode uint64
	avg             map[string]int
	avgEventCode    uint64
}

func (p *personRanks) getSingleEventCount(events []string, max map[string]int) int {
	out := 0
	for _, event := range events {
		if _, ok := p.single[event]; ok {
			out += p.single[event]
		} else {
			out += max[event]
		}
	}
	return out
}

func (p *personRanks) getAvgEventCount(events []string, max map[string]int) int {
	out := 0
	for _, event := range events {
		if _, ok := p.avg[event]; ok {
			out += p.avg[event]
		} else {
			out += max[event]
		}
	}
	return out
}

type eventRanks struct {
	single []*personRanks
	avg    []*personRanks
}

const cutOneEventValue = 3000

func (s *syncer) setStaticDiyEventRanks() (err error) {
	//_ = s.db.AutoMigrate(&types.DiyEventRanks{})
	//_ = s.db.AutoMigrate(&types.DiyEventRanksEventIndex{})
	//s.db.Delete(&types.DiyEventRanks{}, "1 = 1")
	//s.db.Delete(&types.DiyEventRanksEventIndex{}, "1 = 1")
	//defer func() {
	//	if err != nil {
	//		s.db.Delete(&types.DiyEventRanks{}, "1 = 1")
	//		s.db.Delete(&types.DiyEventRanksEventIndex{}, "1 = 1")
	//	}
	//}()

	// 缓存
	var allSingleRank []types.RanksSingle
	var allAvgRank []types.RanksAverage
	s.db.Find(&allSingleRank)
	s.db.Find(&allAvgRank)
	var allPersonRankMap = make(map[string]*personRanks)
	for _, rank := range allSingleRank {
		if _, ok := allPersonRankMap[rank.PersonID]; !ok {
			allPersonRankMap[rank.PersonID] = &personRanks{
				personID:        rank.PersonID,
				single:          make(map[string]int),
				singleEventCode: 0,
				avg:             make(map[string]int),
				avgEventCode:    0,
			}
		}
		allPersonRankMap[rank.PersonID].single[rank.EventID] = rank.WorldRank
		if code, ok := wcaEventMap[rank.EventID]; ok {
			allPersonRankMap[rank.PersonID].singleEventCode |= 1 << code
		}
	}
	for _, rank := range allAvgRank {
		if _, ok := allPersonRankMap[rank.PersonID]; !ok {
			allPersonRankMap[rank.PersonID] = &personRanks{
				personID: rank.PersonID,
				single:   make(map[string]int),
				avg:      make(map[string]int),
			}
		}
		allPersonRankMap[rank.PersonID].avg[rank.EventID] = rank.WorldRank
		if code, ok := wcaEventMap[rank.EventID]; ok {
			allPersonRankMap[rank.PersonID].avgEventCode |= uint64(1) << code
		}
	}

	var maxSingle = make(map[string]int)
	var maxAvg = make(map[string]int)

	// 项目缓存
	var cache = make(map[string]*eventRanks)
	for _, event := range wcaEventsList {
		cache[event] = &eventRanks{
			single: make([]*personRanks, 0),
			avg:    make([]*personRanks, 0),
		}
		for _, person := range allPersonRankMap {
			if _, ok := person.single[event]; ok {
				cache[event].single = append(cache[event].single, person)
			}
			if _, ok := person.avg[event]; ok {
				cache[event].avg = append(cache[event].avg, person)
			}
		}
		maxSingleV := len(cache[event].single) + 1
		maxAvgV := len(cache[event].avg) + 1

		maxSingle[event] = maxSingleV
		maxAvg[event] = maxAvgV
	}

	// 计算
	var idxList []types.DiyEventRanksEventIndex

	//var saveSingleDiyRanks []types.DiyEventSingleRanks
	//var saveAvgDiyRanks []types.DiyEventAvgRanks
	total := 0
	for events := range combinationsStream(wcaEventsList, 2, 12) {
		ts := time.Now()
		// idx
		eventCodeID := encodeEvents(events)
		idx := types.DiyEventRanksEventIndex{
			ID:     eventCodeID,
			Events: strings.Join(events, ","),
		}
		idxList = append(idxList, idx)
		// 累加然后排序

		// 单次
		var singleCache = make([]types.DiyEventSingleRanks, 0, 1024)
		cutValue := cutOneEventValue * len(events)
		for _, person := range allPersonRankMap {
			if !checkHasEvent(eventCodeID, person.singleEventCode) {
				continue
			}

			// 分数不足预算分
			value := person.getSingleEventCount(events, maxSingle)
			if value > cutValue {
				continue
			}

			singleCache = append(singleCache, types.DiyEventSingleRanks{
				DiyEventRanks: types.DiyEventRanks{
					EventIndexID: eventCodeID,
					WcaID:        person.personID,
					Value:        value,
				},
			})
		}
		slices.SortFunc(singleCache, func(a, b types.DiyEventSingleRanks) int {
			return a.Value - b.Value
		})

		total += 1

		n500 := singleCache[len(singleCache)-1].Value
		if len(singleCache) >= 500 {
			n500 = singleCache[499].Value
		}

		fmt.Printf("[%d] [%+v] [%d] \t [%d/%d] | %s\n",
			total, time.Since(ts), len(singleCache), n500, cutValue, strings.Join(events, ","))

		singleCache = nil // GC
	}

	return err
}

//// 跳过平均
//if slices.Contains(events, "333mbf") {
//	continue
//}
//var avgCache []types.DiyEventAvgRanks

func (s *syncer) setStaticDiyEventSingleRanks() (err error) {
	var allSingleRank []types.RanksSingle
	s.db.Find(&allSingleRank)

	allPersonRankMap := make(map[string]*personRanks)

	for _, rank := range allSingleRank {
		if !slices.Contains(wcaEventsList, rank.EventID) {
			continue
		}
		if _, ok := allPersonRankMap[rank.PersonID]; !ok {
			allPersonRankMap[rank.PersonID] = &personRanks{
				personID: rank.PersonID,
				single:   make(map[string]int),
			}
		}
		allPersonRankMap[rank.PersonID].single[rank.EventID] = rank.WorldRank
	}

	// person 索引
	personList := make([]*personRanks, 0, len(allPersonRankMap))
	personIndex := make(map[string]int)
	for _, p := range allPersonRankMap {
		idx := len(personList)
		personList = append(personList, p)
		personIndex[p.personID] = idx
	}

	//  bitset 倒排索引
	eventBitsets := make(map[string]*utils_tool.Bitset)

	for _, e := range wcaEventsList {
		eventBitsets[e] = utils_tool.NewBitset(len(personList))
	}

	for _, p := range personList {
		idx := personIndex[p.personID]
		for e := range p.single {
			eventBitsets[e].Set(idx)
		}
	}

	//  max 值
	maxSingle := make(map[string]int)
	for _, e := range wcaEventsList {
		cnt := 0
		eventBitsets[e].ForEach(func(i int) {
			cnt++
		})
		maxSingle[e] = cnt + 1
	}

	// 主计算
	idx := 0
	for events := range combinationsStream(wcaEventsList, 2, 17) {
		ts := time.Now()
		// bitset 交集
		eventCodeID := encodeEvents(events)
		bs := eventBitsets[events[0]]
		for i := 1; i < len(events); i++ {
			bs = bs.Or(eventBitsets[events[i]])
		}

		cutValue := cutOneEventValue * len(events)
		h := &utils_tool.MaxHeap{}
		heap.Init(h)

		count := 0
		bs.ForEach(func(i int) {
			p := personList[i]

			value := p.getSingleEventCount(events, maxSingle)
			count++
			if value > cutValue {
				return
			}

			r := utils_tool.RankItem{Value: value}
			r.SetData(p)

			if h.Len() < 500 {
				heap.Push(h, r)
			} else if (*h)[0].Value > value {
				(*h)[0] = r
				heap.Fix(h, 0)
			}
		})

		topList := h.Copy()
		slices.SortFunc(topList, func(a, b utils_tool.RankItem) int {
			return a.Value - b.Value
		})

		var data []types.DiyEventSingleRanks
		rank := 1
		for i, item := range topList {
			if i > 0 && item.Value != topList[i-1].Value {
				rank = i + 1
			}
			data = append(data, types.DiyEventSingleRanks{
				DiyEventRanks: types.DiyEventRanks{
					EventIndexID: eventCodeID,
					WcaID:        item.Data.(*personRanks).personID,
					Value:        item.Value,
					Rank:         rank,
					Total:        count,
				},
			})
		}
		if h.Len() < 500 {
			fmt.Printf("[WARN] max heap size: %d, cut: %d, total: %d, events: %s\n", h.Len(), cutValue, count, strings.Join(events, ","))
		}
		n500 := 0
		if h.Len() > 0 {
			n500 = (*h)[0].Value
		}
		idx += 1
		fmt.Printf("[%d] [%+v] [%d] \t [%d/%d] | %s\n",
			idx,
			time.Since(ts),
			count,
			n500,
			cutValue,
			strings.Join(events, ","),
		)
	}
	return nil
}
