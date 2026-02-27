package wca

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/guojia99/cubing-pro/src/wca/types"
	"github.com/guojia99/cubing-pro/src/wca/utils"
)

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

func (w *wca) setCompetitionName(results []types.Result) []types.Result {
	var compID []string

	for _, result := range results {
		compID = append(compID, result.CompetitionID)
	}

	var comps []types.Competition
	w.db.Where("id in ?", compID).Find(&comps)
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
	out = w.setCompetitionName(out)

	return out, nil
}

func (w *wca) GetAllPersons() []types.Person {
	if data, ok := w.cache.Get("GetAllPersons"); ok {
		return data.([]types.Person)
	}

	var out []types.Person
	if err := w.db.Find(&out).Error; err != nil {
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
