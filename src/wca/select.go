package wca

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/guojia99/cubing-pro/src/wca/types"
	"github.com/guojia99/cubing-pro/src/wca/utils"
)

func (w *wca) getRoundTypeMap() map[string]types.RoundType {
	var out = make(map[string]types.RoundType)

	var rounds []types.RoundType
	w.db.Find(&rounds)

	for _, round := range rounds {
		out[round.ID] = round
	}
	return out
}

func (w *wca) getCountryPersons(country string) []types.Person {
	if out, ok := w.cache.Get(country); ok {
		return out.([]types.Person)
	}

	country = w.getCountryID(country)
	query := w.db.Model(&types.Person{})
	if country != "" {
		query.Where("country_id = ?", country)
	}
	query.Where("sub_id = 1")

	var out []types.Person
	query.Find(&out)
	w.cache.Set(country, out, time.Minute*5)
	return out
}

func (w *wca) CountryList() []types.Country {
	var out []types.Country
	w.db.Find(&out)
	return out
}

func (w *wca) GetAllCountry() map[string]types.Country {

	data, ok := w.cache.Get("GetAllCountry")
	if ok {
		return data.(map[string]types.Country)
	}

	var out = make(map[string]types.Country)

	var list []types.Country
	w.db.Find(&list)

	for _, country := range list {
		out[country.ID] = country
	}

	w.cache.Set("GetAllCountry", out, time.Minute*120)
	return out
}

func (w *wca) GetPersonInfo(wcaId string) (types.PersonInfo, error) {
	var person types.Person
	if err := w.db.Where("wca_id = ?", wcaId).First(&person).Error; err != nil {
		return types.PersonInfo{}, fmt.Errorf("not found wca id %s", wcaId)
	}

	var results []types.Result
	if err := w.db.Where("person_id = ?", wcaId).Find(&results).Error; err != nil {
		return types.PersonInfo{}, err
	}

	var bestRanks []types.RanksSingle
	var avgRanks []types.RanksAverage

	if err := w.db.Where("person_id = ?", wcaId).Find(&bestRanks).Error; err != nil {
		return types.PersonInfo{}, err
	}
	if err := w.db.Where("person_id = ?", wcaId).Find(&avgRanks).Error; err != nil {
		return types.PersonInfo{}, err
	}

	var out = types.PersonInfo{
		PersonName:       person.Name,
		WcaID:            person.WcaID,
		CountryID:        person.CountryID,
		Gender:           person.Gender,
		CompetitionCount: 0,
		MedalCount: types.MedalCount{
			Gold:   0,
			Silver: 0,
			Bronze: 0,
			Total:  0,
		},
		RecordCount: types.RecordCount{
			National:    0,
			Continental: 0,
			World:       0,
			Total:       0,
		},
		PersonalRecords: make(map[string]types.PersonalRecord),
	}

	var compsMap = make(map[string]string)
	for _, result := range results {
		compsMap[result.CompetitionID] = result.CompetitionID
		// 记录
		if result.RegionalSingleRecord != "" {
			switch result.RegionalSingleRecord {
			case "WR":
				out.RecordCount.World += 1
			case "NR":
				out.RecordCount.National += 1
			case "AsR", "ER", "NAR", "OcR", "SAR", "AfR":
				out.RecordCount.Continental += 1
			}
			out.RecordCount.Total += 1
		}
		if result.RegionalAverageRecord != "" {
			switch result.RegionalAverageRecord {
			case "WR":
				out.RecordCount.World += 1
			case "NR":
				out.RecordCount.National += 1
			case "AsR", "ER", "NAR", "OcR", "SAR", "AfR":
				out.RecordCount.Continental += 1
			}
			out.RecordCount.Total += 1
		}

		// 排名
		switch result.FormatID {
		case "f", "c":
			if result.Pos <= 3 {
				out.MedalCount.Total += 1
			}
			if result.Pos == 1 {
				out.MedalCount.Gold += 1
			}
			if result.Pos == 2 {
				out.MedalCount.Silver += 1
			}
			if result.Pos == 3 {
				out.MedalCount.Bronze += 1
			}
		}
	}
	out.CompetitionCount = len(compsMap)

	for _, r := range bestRanks {
		if _, ok := out.PersonalRecords[r.EventID]; !ok {
			out.PersonalRecords[r.EventID] = types.PersonalRecord{
				Best: nil,
				Avg:  nil,
			}
		}

		b := &types.PersonResult{
			EventId:       r.EventID,
			Best:          r.Best,
			BestStr:       utils.ResultsTimeFormat(r.Best, r.EventID),
			PersonName:    person.Name,
			PersonId:      r.PersonID,
			WorldRank:     r.WorldRank,
			ContinentRank: r.ContinentRank,
			CountryRank:   r.CountryRank,
		}

		cc := out.PersonalRecords[r.EventID]
		cc.Best = b
		out.PersonalRecords[r.EventID] = cc
	}
	for _, r := range avgRanks {
		if _, ok := out.PersonalRecords[r.EventID]; !ok {
			out.PersonalRecords[r.EventID] = types.PersonalRecord{
				Best: nil,
				Avg:  nil,
			}
		}

		a := &types.PersonResult{
			EventId:       r.EventID,
			Best:          r.Best,
			BestStr:       utils.ResultsTimeFormat(r.Best, r.EventID),
			PersonName:    person.Name,
			PersonId:      r.PersonID,
			WorldRank:     r.WorldRank,
			ContinentRank: r.ContinentRank,
			CountryRank:   r.CountryRank,
		}

		cc := out.PersonalRecords[r.EventID]
		cc.Avg = a
		out.PersonalRecords[r.EventID] = cc
	}

	cts := w.GetAllCountry()

	out.CountryIso2 = cts[person.CountryID].ISO2

	return out, nil
}

func (w *wca) ExportToTable(filePath string) error {
	//TODO implement me
	panic("implement me")
}

func (w *wca) GetCompetition(compId string) (types.Competition, error) {
	//TODO implement me
	panic("implement me")
}

func (w *wca) GetPersonCompetition(wcaId string) ([]types.Competition, error) {
	var out []types.Result

	if err := w.db.Where("person_id = ?", wcaId).Find(&out).Error; err != nil {
		return nil, err
	}

	compIdsMap := make(map[string]bool)
	var compIds []string
	for _, result := range out {
		if result.CompetitionID != "" {
			if !compIdsMap[result.CompetitionID] {
				compIdsMap[result.CompetitionID] = true
				compIds = append(compIds, result.CompetitionID)
			}
		}
	}

	var comps []types.Competition
	if err := w.db.Where("id IN (?)", compIds).Find(&comps).Error; err != nil {
		return nil, err
	}

	var country []types.Country
	w.db.Find(&country)
	var countryMap = make(map[string]types.Country)
	for _, r := range country {
		countryMap[r.Name] = r
	}

	for idx := 0; idx < len(comps); idx++ {
		cp := comps[idx]
		comps[idx].CountryIso2 = countryMap[comps[idx].CountryID].ISO2
		comps[idx].EventIds = strings.Split(comps[idx].EventSpecs, " ")
		comps[idx].StartDate = time.Date(int(cp.Year), time.Month(cp.Month), int(cp.Day), 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		comps[idx].EndDate = time.Date(int(cp.EndYear), time.Month(cp.EndMonth), int(cp.EndDay), 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	}

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

func (w *wca) getResultAttemptMap(results []types.Result) map[int64][]int64 {
	var resultIDs []int64
	for _, result := range results {
		resultIDs = append(resultIDs, result.ID)
	}

	var resultAttemptMap = make(map[int64][]int64)
	for i := 0; i < len(resultIDs); i += 5000 {
		end := i + 5000
		if end > len(resultIDs) {
			end = len(resultIDs)
		}
		batchIDs := resultIDs[i:end]

		var batchAtt []types.ResultAttempt
		if err := w.db.Where("result_id IN ?", batchIDs).Find(&batchAtt).Error; err != nil {
			continue
		}
		sort.Slice(batchAtt, func(i, j int) bool {
			return batchAtt[i].AttemptNumber <= batchAtt[j].AttemptNumber
		})

		for _, att := range batchAtt {
			if _, ok := resultAttemptMap[att.ResultID]; !ok {
				resultAttemptMap[att.ResultID] = make([]int64, 0)
			}
			resultAttemptMap[att.ResultID] = append(resultAttemptMap[att.ResultID], att.Value)
		}
	}

	return resultAttemptMap
}

func (w *wca) setResultAttempts(results []types.Result) []types.Result {
	resultAttemptMap := w.getResultAttemptMap(results)

	for idx := 0; idx < len(results); idx++ {
		r := results[idx]
		attempts := resultAttemptMap[r.ID]

		results[idx].BestIndex = -1
		results[idx].WorstIndex = -1
		results[idx].Attempts = attempts

		if len(results[idx].Attempts) != 5 {
			continue
		}

		// 1. 找 BestIndex：最小的正数
		bestIndex := -1
		for i, v := range attempts {
			if v > 0 {
				if bestIndex == -1 || v < attempts[bestIndex] {
					bestIndex = i
				}
			}
		}

		// 2. 找 WorstIndex：优先负数，否则最大正数
		worstIndex := -1
		for i, v := range attempts {
			if worstIndex == -1 {
				worstIndex = i
				continue
			}

			curr := attempts[worstIndex]
			// 如果当前 v 是无效（<0），而 curr 是有效（>=0），v 更差
			if v < 0 && curr >= 0 {
				worstIndex = i
			} else if v >= 0 && curr < 0 {
				// curr 已经是无效，v 是有效，不换
				continue
			} else if v < 0 {

			} else {
				// 都是有效，选更大的（更慢）
				if v > curr {
					worstIndex = i
				}
			}
		}
		results[idx].BestIndex = bestIndex   // 可能为 -1（全无效）
		results[idx].WorstIndex = worstIndex // 至少有一个元素，不会为 -1
	}
	return results
}

func (w *wca) setCompetitionNameAndSort(results []types.Result) []types.Result {
	var compID []string
	var compMap = make(map[string]types.Competition)

	for _, result := range results {
		compID = append(compID, result.CompetitionID)
	}

	var comps []types.Competition
	w.db.Where("id in ?", compID).Find(&comps)
	for _, comp := range comps {
		compMap[comp.ID] = comp
	}

	// 填充名字
	var mp = make(map[string]string)
	var tmp = make(map[string]string)
	for _, comp := range comps {
		mp[comp.ID] = comp.Name
		tmp[comp.ID] = fmt.Sprintf("%d-%d-%d", comp.Year, int(comp.Month), int(comp.Day))
	}

	for idx := range results {
		results[idx].CompetitionName = mp[results[idx].CompetitionID]
		results[idx].CompetitionTime = tmp[results[idx].CompetitionID]
	}

	eventOrder := buildEventOrderMap()
	roundOrder := w.getRoundTypeMap()
	// 基于比赛时间排序和项目进行排序
	sort.Slice(results, func(i, j int) bool {
		a, b := results[i], results[j]

		// 比赛时间排序
		compA := compMap[a.CompetitionID]
		compB := compMap[b.CompetitionID]

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

		//项目排序
		orderA := eventOrder[a.EventID]
		orderB := eventOrder[b.EventID]
		if _, exists := eventOrder[a.EventID]; !exists {
			orderA = len(wcaEventsList)
		}
		if _, exists := eventOrder[b.EventID]; !exists {
			orderB = len(wcaEventsList)
		}
		if orderA != orderB {
			return orderA < orderB
		}

		// 项目轮次排序
		roundA := roundOrder[a.RoundTypeID]
		roundB := roundOrder[b.RoundTypeID]
		if roundA != roundB {
			return roundA.Rank < roundB.Rank
		}

		// 最后ID排序
		return a.ID < b.ID
	})
	return results
}

func (w *wca) GetPersonResult(wcaId string) ([]types.Result, error) {
	var out []types.Result

	if err := w.db.Where("person_id = ?", wcaId).Find(&out).Error; err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return nil, nil
	}
	out = w.setResultAttempts(out)
	out = w.setCompetitionNameAndSort(out)

	return out, nil
}

func (w *wca) GetAllPersons() []types.Person {
	if data, ok := w.cache.Get("GetAllPersons"); ok {
		return data.([]types.Person)
	}

	var out []types.Person
	if err := w.db.Where("sub_id = 1").Find(&out).Error; err != nil {
		return nil
	}

	country := w.GetAllCountry()
	for idx := 0; idx < len(out); idx++ {
		out[idx].Iso2 = country[out[idx].CountryID].ISO2
	}
	w.cache.Set("GetAllPersons", out, time.Hour)
	return out
}

func (w *wca) SearchPlayers(query string) []types.Person {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}

	allPersons := w.GetAllPersons()
	if allPersons == nil {
		return nil
	}

	wcaQuery := strings.ToUpper(query)
	var out []types.Person
	for _, p := range allPersons {
		if strings.Contains(p.WcaID, wcaQuery) {
			out = append(out, p)
		}
		if strings.Contains(p.Name, query) {
			out = append(out, p)
		}
	}
	return out
}

func (w *wca) GetGrandSlam() []types.AllEventChampionshipsPodium {
	var out []types.AllEventChampionshipsPodium
	w.db.Find(&out)
	return out
}

func rankWithEventsPaginate(fullList []types.RankWithEventsStatic, page, size int) ([]types.RankWithEventsStatic, int64) {
	count := int64(len(fullList))
	start := (page - 1) * size
	if start < 0 {
		start = 0
	}
	if start >= len(fullList) {
		return nil, count
	}
	end := start + size
	if end > len(fullList) {
		end = len(fullList)
	}
	return fullList[start:end], count
}

type rankEntry struct {
	PersonID string
	EventID  string
	Rank     int
}

func (w *wca) rankWithEventsLoadRanks(wcaIDs, events []string, avg bool, useWorldRank bool) ([]rankEntry, error) {
	eventSet := make(map[string]bool)
	for _, e := range events {
		eventSet[e] = true
	}
	getRank := func(world, country int) int {
		if useWorldRank {
			return world
		}
		return country
	}

	var entries []rankEntry
	addFromRows := func(rows interface{}) {
		switch r := rows.(type) {
		case []types.RanksAverage:
			for _, x := range r {
				if rank := getRank(x.WorldRank, x.CountryRank); rank > 0 && eventSet[x.EventID] {
					entries = append(entries, rankEntry{x.PersonID, x.EventID, rank})
				}
			}
		case []types.RanksSingle:
			for _, x := range r {
				if rank := getRank(x.WorldRank, x.CountryRank); rank > 0 && eventSet[x.EventID] {
					entries = append(entries, rankEntry{x.PersonID, x.EventID, rank})
				}
			}
		}
	}

	for i := 0; i < len(wcaIDs); i += 5000 {
		end := i + 5000
		if end > len(wcaIDs) {
			end = len(wcaIDs)
		}
		batch := wcaIDs[i:end]
		if avg {
			var rows []types.RanksAverage
			if err := w.db.Where("person_id IN ? AND event_id IN ?", batch, events).Find(&rows).Error; err != nil {
				return nil, err
			}
			addFromRows(rows)
		} else {
			var rows []types.RanksSingle
			if err := w.db.Where("person_id IN ? AND event_id IN ?", batch, events).Find(&rows).Error; err != nil {
				return nil, err
			}
			addFromRows(rows)
		}
	}
	return entries, nil
}

func (w *wca) getRankWithEventsFullList(events []string, country string, avg bool) (out []types.RankWithEventsStatic, err error) {
	// events 为空则使用全项目
	if len(events) == 0 {
		events = make([]string, len(wcaEventsList))
		copy(events, wcaEventsList)
	}
	// 使用平均时过滤掉 333mbf（无平均）
	if avg {
		filtered := make([]string, 0, len(events))
		for _, e := range events {
			if e != "333mbf" {
				filtered = append(filtered, e)
			}
		}
		events = filtered
	}
	if len(events) == 0 {
		return nil, nil
	}

	eventsKey := make([]string, len(events))
	copy(eventsKey, events)
	sort.Strings(eventsKey)

	useWorldRank := country == ""
	persons := w.getCountryPersons(country)
	if len(persons) == 0 {
		return nil, nil
	}

	wcaIDSet := make(map[string]types.Person)
	for _, p := range persons {
		wcaIDSet[p.WcaID] = p
	}
	wcaIDs := make([]string, 0, len(wcaIDSet))
	for id := range wcaIDSet {
		wcaIDs = append(wcaIDs, id)
	}

	entries, err := w.rankWithEventsLoadRanks(wcaIDs, events, avg, useWorldRank)
	if err != nil {
		return nil, err
	}

	personEventRank := make(map[string]map[string]int)
	for _, e := range entries {
		if personEventRank[e.PersonID] == nil {
			personEventRank[e.PersonID] = make(map[string]int)
		}
		personEventRank[e.PersonID][e.EventID] = e.Rank
	}

	defaultRank := make(map[string]int)
	for _, evt := range events {
		cnt := 0
		for _, m := range personEventRank {
			if _, ok := m[evt]; ok {
				cnt++
			}
		}
		defaultRank[evt] = cnt + 1
	}

	participantSums := make(map[string]int)
	for personID, eventRanks := range personEventRank {
		hasAtLeastOne := false
		sum := 0
		for _, evt := range events {
			if r, ok := eventRanks[evt]; ok {
				hasAtLeastOne = true
				sum += r
			} else {
				sum += defaultRank[evt]
			}
		}
		if hasAtLeastOne {
			participantSums[personID] = sum
		}
	}

	type item struct {
		wcaID string
		name  string
		count int
	}
	var items []item
	for personID, sum := range participantSums {
		if p, ok := wcaIDSet[personID]; ok {
			items = append(items, item{p.WcaID, p.Name, sum})
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].count < items[j].count })

	fullList := make([]types.RankWithEventsStatic, 0, len(items))
	rank := 1
	for i := 0; i < len(items); i++ {
		if i > 0 && items[i].count != items[i-1].count {
			rank = i + 1
		}
		fullList = append(fullList, types.RankWithEventsStatic{
			WcaID: items[i].wcaID,
			Name:  items[i].name,
			Rank:  rank,
			Count: items[i].count,
		})
	}
	return fullList, nil
}

func (w *wca) GetRankWithEvents(events []string, country string, avg bool, page int, size int) (out []types.RankWithEventsStatic, count int64, err error) {
	fullList, err := w.getRankWithEventsFullList(events, country, avg)
	if err != nil {
		return nil, 0, err
	}
	if len(fullList) == 0 {
		return nil, 0, nil
	}
	out, count = rankWithEventsPaginate(fullList, page, size)
	return out, count, nil
}

type personRank struct {
	singles []types.RanksSingle
	avgs    []types.RanksAverage
}

func (w *wca) cacheSingleAndAverageRank() map[string]*personRank {
	key := "_cacheSingleAndAverageRank_cache"
	if out, ok := w.cache.Get(key); ok {
		return out.(map[string]*personRank)
	}

	var singles []types.RanksSingle
	var avgs []types.RanksAverage
	w.db.Find(&singles)
	w.db.Find(&avgs)

	var out = make(map[string]*personRank)

	for _, single := range singles {
		if _, ok := out[single.PersonID]; !ok {
			out[single.PersonID] = &personRank{
				singles: make([]types.RanksSingle, 0),
				avgs:    make([]types.RanksAverage, 0),
			}
		}

		out[single.PersonID].singles = append(out[single.PersonID].singles, single)
	}

	for _, avg := range avgs {
		if _, ok := out[avg.PersonID]; !ok {
			out[avg.PersonID] = &personRank{
				avgs:    make([]types.RanksAverage, 0),
				singles: make([]types.RanksSingle, 0),
			}
		}
		out[avg.PersonID].avgs = append(out[avg.PersonID].avgs, avg)
	}
	w.cache.Set(key, out, time.Minute*30)
	return out
}

func (w *wca) getSingleAndAverageRanks(country string) ([]types.RanksSingle, []types.RanksAverage) {
	// 只查询一次：person 列表、singles、avgs
	persons := w.getCountryPersons(country)

	if len(persons) == 0 {
		return nil, nil
	}

	var singles []types.RanksSingle
	var avgs []types.RanksAverage

	cache := w.cacheSingleAndAverageRank()
	for _, person := range persons {
		pp, ok := cache[person.WcaID]
		if !ok {
			continue
		}

		singles = append(singles, pp.singles...)
		avgs = append(avgs, pp.avgs...)
	}
	return singles, avgs
}

//// getCountryBestWithEventGroupRankOnlyCountry 国家现场算
//func (w *wca) getCountryBestWithEventGroupRankOnlyCountry()

// GetCountryBestWithEventGroupRank 获取某个选手最佳的排列组合
func (w *wca) GetCountryBestWithEventGroupRank(wcaId string, avg bool, useWorld bool) (out []types.RankWithEventsGrouptatic, err error) {
	var person types.Person
	if err = w.db.Where("wca_id = ?", wcaId).Where("sub_id = 1").First(&person).Error; err != nil {
		return nil, err
	}
	country := ""
	if !useWorld {
		country = person.CountryID
	}
	singles, avgs := w.getSingleAndAverageRanks(country)

	// 构建 personID -> eventID -> rank
	getRank := func(world, country int) int {
		if useWorld {
			return world
		}
		return country
	}
	globalData := make(map[string]map[string]int)
	if avg {
		for _, r := range avgs {
			if r.EventID == "333mbf" {
				continue
			}
			rankVal := getRank(r.WorldRank, r.CountryRank)
			if rankVal <= 0 {
				continue
			}
			if globalData[r.PersonID] == nil {
				globalData[r.PersonID] = make(map[string]int)
			}
			globalData[r.PersonID][r.EventID] = rankVal
		}
	} else {
		for _, r := range singles {
			rankVal := getRank(r.WorldRank, r.CountryRank)
			if rankVal <= 0 {
				continue
			}
			if globalData[r.PersonID] == nil {
				globalData[r.PersonID] = make(map[string]int)
			}
			globalData[r.PersonID][r.EventID] = rankVal
		}
	}

	targetEventsMap, ok := globalData[wcaId]
	if !ok || len(targetEventsMap) == 0 {
		return nil, nil
	}

	// 按 wcaEventsList 顺序得到 validEvents
	validEvents := make([]string, 0, len(targetEventsMap))
	for _, evt := range wcaEventsList {
		if _, exists := targetEventsMap[evt]; exists {
			validEvents = append(validEvents, evt)
		}
	}
	n := len(validEvents)
	if n == 0 {
		return nil, nil
	}

	// 每个项目的默认排名：该项目有成绩人数 + 1
	defaultRank := make(map[string]int)
	for _, evt := range validEvents {
		cnt := 0
		for _, m := range globalData {
			if _, exists := m[evt]; exists {
				cnt++
			}
		}
		defaultRank[evt] = cnt + 1
	}

	// 按 排名/总人数 升序排序，比值最小的项目优先计算
	sort.Slice(validEvents, func(i, j int) bool {
		r1, r2 := targetEventsMap[validEvents[i]], targetEventsMap[validEvents[j]]
		t1 := defaultRank[validEvents[i]] - 1
		t2 := defaultRank[validEvents[j]] - 1
		if t1 < 1 {
			t1 = 1
		}
		if t2 < 1 {
			t2 = 1
		}
		return float64(r1)/float64(t1) < float64(r2)/float64(t2)
	})

	// 预计算：personIdx -> eventIdx -> rank，避免内层循环 map 查找
	personIndices := make([]string, 0, len(globalData)-1)
	for pid := range globalData {
		if pid != wcaId {
			personIndices = append(personIndices, pid)
		}
	}
	m := len(personIndices)
	defaultRanks := make([]int, n)
	for i, evt := range validEvents {
		defaultRanks[i] = defaultRank[evt]
	}
	rankMatrix := make([][]int, m)
	hasRank := make([][]bool, m) // 该人在该项目是否有成绩
	for i, pid := range personIndices {
		rankMatrix[i] = make([]int, n)
		hasRank[i] = make([]bool, n)
		for j, evt := range validEvents {
			if r, ok := globalData[pid][evt]; ok {
				rankMatrix[i][j] = r
				hasRank[i][j] = true
			} else {
				rankMatrix[i][j] = defaultRanks[j]
			}
		}
	}
	targetRanks := make([]int, n)
	for j, evt := range validEvents {
		targetRanks[j] = targetEventsMap[evt]
	}

	// 组合迭代器：只枚举 1~6 个项目的组合，剪枝掉 2^n 中大部分无效 mask（约 6 倍减少）
	results := make([]struct {
		events       []string
		sumRank      int
		globalRank   int
		totalPlayers int
	}, 0, 20000)

	var maxEvents = 8
	if useWorld {
		maxEvents = 5
	}

	indices := make([]int, maxEvents)

	// 核心计算
	for k := 1; k <= maxEvents && k <= n; k++ {
		for i := 0; i < k; i++ {
			indices[i] = i
		}
		for {
			targetSum := 0
			currentEvents := make([]string, 0, k)
			for i := 0; i < k; i++ {
				idx := indices[i]
				targetSum += targetRanks[idx]
				currentEvents = append(currentEvents, validEvents[idx])
			}

			betterCount := 0
			totalQualified := 0
			for i := 0; i < m; i++ {
				curHasRank := false
				otherSum := 0
				for j := 0; j < k; j++ {
					evtIdx := indices[j]
					if hasRank[i][evtIdx] {
						curHasRank = true
					}
					otherSum += rankMatrix[i][evtIdx]
					if curHasRank && otherSum >= targetSum {
						break // 剪枝
					}
				}
				if !curHasRank {
					continue
				}
				totalQualified++
				if otherSum < targetSum {
					betterCount++
				}
			}

			results = append(results, struct {
				events       []string
				sumRank      int
				globalRank   int
				totalPlayers int
			}{
				events:       currentEvents,
				sumRank:      targetSum,
				globalRank:   betterCount + 1,
				totalPlayers: totalQualified + 1,
			})

			// 下一个 k 组合
			ii := k - 1
			for ii >= 0 && indices[ii] == n-k+ii {
				ii--
			}
			if ii < 0 {
				break
			}
			indices[ii]++
			for j := ii + 1; j < k; j++ {
				indices[j] = indices[j-1] + 1
			}
		}
	}

	// 排序：GlobalRank 优先，其次 SumRank，其次 event 数量
	sort.Slice(results, func(i, j int) bool {
		if results[i].globalRank != results[j].globalRank {
			return results[i].globalRank < results[j].globalRank
		}
		if results[i].sumRank != results[j].sumRank {
			return results[i].sumRank < results[j].sumRank
		}
		return len(results[i].events) < len(results[j].events)
	})

	// 前 10，不足则全部
	topN := 20
	if len(results) < topN {
		topN = len(results)
	}
	results = results[:topN]

	out = make([]types.RankWithEventsGrouptatic, 0, topN)
	for _, r := range results {
		out = append(out, types.RankWithEventsGrouptatic{
			WcaID:  person.WcaID,
			Name:   person.Name,
			Events: r.events,
			Rank:   r.globalRank,
			Count:  r.sumRank,
		})
	}
	return out, nil
}
