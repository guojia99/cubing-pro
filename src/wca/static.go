package wca

import (
	"sort"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/wca/utils"
	"github.com/guojia99/cubing-pro/src/wca/types"
)

func (w *wca) GetPersonRankTimer(wcaId string) ([]types.StaticWithTimerRank, error) {
	var out []types.StaticWithTimerRank
	if err := w.db.Where("wca_id = ?", wcaId).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func getMaxMonth(year int) int {
	now := time.Now()
	currentYear := now.Year()
	currentMonth := now.Month() // time.Month 是从 1（January）到 12（December）

	if year == currentYear {
		// 如果是今年，则 maxMonth 是上个月
		if currentMonth == time.January {
			// 特殊情况：当前是1月，上个月是去年的12月
			return 12
		}
		return int(currentMonth - 1)
	}
	// 默认返回12
	return 12
}

func (w *wca) getCountryID(country string) string {
	if country == "" {
		return ""
	}
	var dbCountry types.Country
	if err := w.db.Where("iso2 = ?", country).First(&dbCountry).Error; err != nil {
		return country
	}
	country = dbCountry.ID
	return country
}

func (w *wca) GetEventRankWithTimer(eventId, country string, year int, isAvg bool, page, size int) ([]types.StaticWithTimerRank, int64, error) {
	maxMonth := getMaxMonth(year) // 默认值

	// 然后查询该年该月的所有记录
	var results []types.StaticWithTimerRank

	query := w.db.Model(&types.StaticWithTimerRank{}).
		Where("event_id = ? AND year = ? AND month = ?", eventId, year, maxMonth)

	if country != "" {
		country = w.getCountryID(country)
		query = query.Where("country = ?", country)
	}

	// 根据排名类型选择排序字段
	if isAvg {
		query = query.Where("avg_country_rank != 0")
		if country != "" {
			query = query.Order("avg_country_rank ASC") // 排名越小越好，所以升序
		} else {
			query = query.Order("avg_world_rank ASC") // 排名越小越好，所以升序
		}
	} else {
		query = query.Where("single_country_rank != 0")
		if country != "" {
			query = query.Order("single_country_rank ASC") // 排名越小越好，所以升序
		} else {
			query = query.Order("single_world_rank ASC") // 排名越小越好，所以升序
		}
	}

	// 计算总数用于分页
	var total int64
	//countQuery := w.db.Model(&types.StaticWithTimerRank{}).
	//	Where("event_id = ? AND year = ? AND month = ?", eventId, year, maxMonth)
	//
	//if country != "" {
	//	countQuery = countQuery.Where("country = ?", country)
	//}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	// 分页查询
	if page <= 0 {
		page = 1
	}
	if size >= 100 || size <= 0 {
		size = 100
	}

	offset := (page - 1) * size
	err = query.Offset(offset).Limit(size).Find(&results).Error
	if err != nil {
		return nil, 0, err
	}

	// 填充WCAID
	var wcaIDs []string
	for _, result := range results {
		wcaIDs = append(wcaIDs, result.WcaID)
	}
	var ps []types.Person
	w.db.Where("wca_id in (?)", wcaIDs).Find(&ps)
	var personMap = make(map[string]types.Person)
	for _, person := range ps {
		personMap[person.WcaID] = person
	}

	for idx := range len(results) {
		results[idx].WcaName = personMap[results[idx].WcaID].Name
	}

	return results, total, nil
}

func (w *wca) GetEventRankWithFullNow(eventId, country string, isAvg bool, page, size int) ([]types.Result, int64, error) {
	query := w.db.Model(&types.Result{}).Where("event_id = ?", eventId)
	if country != "" {
		country = w.getCountryID(country)
		query = query.Where("person_country_id = ?", country)
	}

	if size > 100 {
		size = 100
	}

	var pagedResults []types.Result
	var total int64
	// 分项目, 333mbf全部查
	if eventId == "333mbf" {
		var results []types.Result
		query.Find(&results)
		sort.Slice(results, func(i, j int) bool {
			a := results[i].Best
			b := results[j].Best
			return utils.IsBestResult(eventId, a, b)
		})
		total = int64(len(results))
		start := (page - 1) * size
		end := start + size
		if start > int(total) {
			start = int(total)
		}
		if end > int(total) {
			end = int(total)
		}
		pagedResults = results[start:end]
	} else {
		if isAvg {
			query = query.Where("average > ?", 0)
			query = query.Order("average")
		} else {
			query = query.Where("best > ?", 0)
			query = query.Order("best")
		}
		query.Count(&total)
		query = query.Offset((page - 1) * size).Limit(size)
		query.Find(&pagedResults)
	}

	// 拓展参数
	pagedResults = w.setResultAttempts(pagedResults)
	pagedResults = w.setCompetitionName(pagedResults)

	return pagedResults, total, nil
}
