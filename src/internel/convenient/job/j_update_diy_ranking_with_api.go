package job

import (
	"sort"
	"strings"
	"time"

	wca_model "github.com/guojia99/cubing-pro/src/internel/database/model/wca"
	"github.com/guojia99/cubing-pro/src/internel/database/model/wca/utils"
	utils2 "github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/wca/types"
)

func PersonBeWcaDBToCubingProDB(in types.PersonInfo) wca_model.PersonBestResults {
	out := wca_model.PersonBestResults{
		DBVersion:        time.Now().Format(time.DateTime),
		PersonName:       in.PersonName,
		WCAID:            in.WcaID,
		Best:             make(map[string]wca_model.Results),
		Avg:              make(map[string]wca_model.Results),
		CompetitionCount: in.CompetitionCount,
		MedalCount: wca_model.MedalCount{
			Gold:   in.MedalCount.Gold,
			Silver: in.MedalCount.Silver,
			Bronze: in.MedalCount.Bronze,
			Total:  in.MedalCount.Total,
		},
		RecordCount: wca_model.RecordCount{
			National:    in.RecordCount.National,
			Continental: in.RecordCount.Continental,
			World:       in.RecordCount.World,
			Total:       in.RecordCount.Total,
		},
	}

	for ev, rcs := range in.PersonalRecords {
		if rcs.Best != nil {
			out.Best[ev] = wca_model.Results{
				EventId:       ev,
				Best:          rcs.Best.Best,
				BestStr:       rcs.Best.BestStr,
				PersonName:    rcs.Best.PersonName,
				PersonId:      rcs.Best.PersonId,
				WorldRank:     rcs.Best.WorldRank,
				ContinentRank: rcs.Best.ContinentRank,
				CountryRank:   rcs.Best.CountryRank,
				Rank:          rcs.Best.Rank,
			}
		}
		if rcs.Avg != nil {
			out.Avg[ev] = wca_model.Results{
				EventId:       ev,
				Average:       rcs.Avg.Best,
				AverageStr:    rcs.Avg.BestStr,
				PersonName:    rcs.Best.PersonName,
				PersonId:      rcs.Best.PersonId,
				WorldRank:     rcs.Best.WorldRank,
				ContinentRank: rcs.Best.ContinentRank,
				CountryRank:   rcs.Best.CountryRank,
				Rank:          rcs.Best.Rank,
			}
		}
	}

	return out
}

func (u *UpdateDiyRankings) apiGetAllResult(WcaIDs []string) map[string]types.PersonInfo {
	var out = make(map[string]types.PersonInfo)

	WcaIDs = utils2.RemoveRepeatedElement(WcaIDs)

	var resultsCh []types.PersonInfo

	for _, wcaId := range WcaIDs {
		wcaId = strings.ToUpper(wcaId)
		if len(wcaId) != 10 {
			continue
		}
		//log.Printf("[apiGetAllResult] %+v\n", wcaId)
		res, err := u.Wca.GetPersonInfo(wcaId)
		//res, err := wca_api.GetWcaResultWithDbAndAPI(u.DB, wcaId)
		if err != nil {
			//log.Printf("[apiGetAllResult] get wca %s error %+v\n", wcaId, err)
			continue
		}

		// 缓存到数据库
		var dbResult wca_model.WCAResult
		if err = u.DB.Where("wca_id = ?", wcaId).First(&dbResult).Error; err == nil {
			dbResult.PersonBestResults = PersonBeWcaDBToCubingProDB(res)
		} else {
			dbResult = wca_model.WCAResult{
				WcaID:             wcaId,
				PersonBestResults: PersonBeWcaDBToCubingProDB(res),
			}
		}
		u.DB.Save(&dbResult)

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
