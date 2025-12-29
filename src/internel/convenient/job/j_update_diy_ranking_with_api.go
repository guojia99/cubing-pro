package job

import (
	"log"
	"sort"
	"strings"

	"github.com/guojia99/cubing-pro/src/internel/database/model/wca/utils"
	utils2 "github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/wca/types"
)

func (u *UpdateDiyRankings) apiGetAllResult(WcaIDs []string) map[string]types.PersonInfo {
	var out = make(map[string]types.PersonInfo)

	WcaIDs = utils2.RemoveRepeatedElement(WcaIDs)

	var resultsCh []types.PersonInfo

	for _, wcaId := range WcaIDs {
		wcaId = strings.ToUpper(wcaId)
		if len(wcaId) != 10 {
			continue
		}
		log.Printf("[apiGetAllResult] %+v\n", wcaId)
		res, err := u.Wca.GetPersonInfo(wcaId)
		//res, err := wca_api.GetWcaResultWithDbAndAPI(u.DB, wcaId)
		if err != nil {
			log.Printf("[apiGetAllResult] get wca %s error %+v\n", wcaId, err)
			continue
		}

		resultsCh = append(resultsCh, res)
	}

	for _, res := range resultsCh {
		out[res.PersonName] = res
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
		var bests []types.PersonResult
		var avgs []types.PersonResult

		for _, r := range data {
			if _, ok := r.PersonalRecords[eid]; !ok {
				continue
			}

			if r.PersonalRecords[eid].Best != nil {
				bests = append(bests, *r.PersonalRecords[eid].Best)
			}
			if r.PersonalRecords[eid].Avg != nil {
				avgs = append(avgs, *r.PersonalRecords[eid].Avg)
			}
		}

		sort.Slice(bests, func(i, j int) bool {
			return utils.IsBestResult(bests[i].EventId, bests[i].Best, bests[j].Best)
		})

		sort.Slice(avgs, func(i, j int) bool {
			return utils.IsBestResult(avgs[i].EventId, avgs[i].Best, avgs[j].Best)
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

			if idx >= 1 && wrs[idx-1].AvgStr == a.BestStr {
				index = wrs[idx-1].AvgRank
			}
			wrs[idx].AvgRank = index
			wrs[idx].AvgStr = a.BestStr
			wrs[idx].AvgPersonName = a.PersonName
			wrs[idx].AvgPersonWCAID = a.PersonId
		}
		out[eid] = wrs
	}
	return out
}
