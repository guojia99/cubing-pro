package job

import (
	"log"
	"sort"
	"time"

	utils2 "github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/internel/wca"

	wca2 "github.com/guojia99/cubing-pro/src/internel/database/model/wca"
	"github.com/guojia99/cubing-pro/src/internel/database/wca_model/utils"
)

func (u *UpdateDiyRankings) getWcaResultWithDbAndAPI(wcaId string) (*wca.PersonBestResults, error) {
	// 从db中查询
	var dbResult wca2.WCAResult
	if err := u.DB.Where("wca_id = ?", wcaId).First(&dbResult).Error; err == nil {
		if time.Since(dbResult.UpdatedAt) <= time.Hour*6 {
			return &dbResult.PersonBestResults, nil
		}
	}

	// api真实查询
	time.Sleep(time.Second)
	res, err := wca.ApiGetWCAResults(wcaId)
	if err != nil {
		return nil, err
	}

	// 缓存到数据库
	if dbResult.ID != 0 {
		dbResult.PersonBestResults = *res
		_ = u.DB.Save(&dbResult)
		return res, nil
	}
	dbResult = wca2.WCAResult{
		WcaID:             wcaId,
		PersonBestResults: *res,
	}
	_ = u.DB.Create(&dbResult)
	return res, nil
}

func (u *UpdateDiyRankings) apiGetAllResult(WcaIDs []string) map[string]wca.PersonBestResults {
	var out = make(map[string]wca.PersonBestResults)

	WcaIDs = utils2.RemoveRepeatedElement(WcaIDs)

	var resultsCh []*wca.PersonBestResults

	for _, wcaId := range WcaIDs {
		if len(wcaId) != 10 {
			continue
		}
		log.Printf("[apiGetAllResult] %+v\n", wcaId)
		res, err := u.getWcaResultWithDbAndAPI(wcaId)
		if err != nil {
			log.Printf("[apiGetAllResult] get wca %s error %+v\n", wcaId, err)
			continue
		}

		resultsCh = append(resultsCh, res)
	}

	for _, res := range resultsCh {
		out[res.PersonName] = *res
	}
	return out
}

var wcaEventsList = []string{
	"333",
	"222",
	"444",
	"555",
	"666",
	"777",
	"333bf",
	"333fm",
	"333oh",
	"clock",
	"minx",
	"pyram",
	"skewb",
	"sq1",
	"444bf",
	"555bf",
	"333mbf",
}

func (u *UpdateDiyRankings) apiGetSortResult(WcaIDs []string) map[string][]WcaResult {
	var out = make(map[string][]WcaResult)
	data := u.apiGetAllResult(WcaIDs)

	for _, eid := range wcaEventsList {
		var bests []wca.Results
		var avgs []wca.Results

		for _, r := range data {
			if b, ok := r.Best[eid]; ok {
				bests = append(bests, b)
			}
			if a, ok := r.Avg[eid]; ok {
				avgs = append(avgs, a)
			}
		}

		sort.Slice(bests, func(i, j int) bool {
			return utils.IsBestResult(bests[i].EventId, bests[i].Best, bests[j].Best)
		})

		sort.Slice(avgs, func(i, j int) bool {
			return utils.IsBestResult(avgs[i].EventId, avgs[i].Average, avgs[j].Average)
		})
		var wrs []WcaResult
		for idx, b := range bests {
			var index = idx + 1
			if idx >= 1 && wrs[idx-1].BestStr == b.BestStr {
				index = wrs[idx-1].BestRank
			}
			wrs = append(
				wrs, WcaResult{
					BestRank:        index,
					BestStr:         b.BestStr,
					BestPersonName:  b.PersonName,
					BestPersonWCAID: b.PersonId,
				},
			)
		}
		for idx, a := range avgs {
			var index = idx + 1

			if idx >= 1 && wrs[idx-1].AvgStr == a.AverageStr {
				index = wrs[idx-1].AvgRank
			}
			wrs[idx].AvgRank = index
			wrs[idx].AvgStr = a.AverageStr
			wrs[idx].AvgPersonName = a.PersonName
			wrs[idx].AvgPersonWCAID = a.PersonId
		}
		out[eid] = wrs
	}
	return out
}
