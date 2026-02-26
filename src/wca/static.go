package wca

import (
	"github.com/guojia99/cubing-pro/src/wca/types"
)

func (w *wca) GetPersonRankTimer(wcaId string) ([]types.StaticWithTimerRank, error) {
	var out []types.StaticWithTimerRank
	if err := w.db.Where("wca_id = ?", wcaId).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wca) GetEventRankWithTimer(eventId, country string, year int, isAvg bool, page, size int) ([]types.StaticWithTimerRank, int64, error) {
	maxMonth := 12 // 默认值

	// 然后查询该年该月的所有记录
	var results []types.StaticWithTimerRank

	query := w.db.Model(&types.StaticWithTimerRank{}).
		Where("event_id = ? AND year = ? AND month = ?", eventId, year, maxMonth)

	if country != "" {
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
