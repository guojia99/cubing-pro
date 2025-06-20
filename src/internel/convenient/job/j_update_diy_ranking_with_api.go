package job

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	utils2 "github.com/guojia99/cubing-pro/src/internel/utils"

	"github.com/guojia99/cubing-pro/src/internel/database/wca_model/utils"
)

type WCAResults struct {
	Id            int    `json:"id"`
	Best          int    `json:"best"`
	Average       int    `json:"average"`
	Name          string `json:"name"`
	CompetitionId string `json:"competition_id"`
	EventId       string `json:"event_id"`
	WcaId         string `json:"wca_id"`
	Attempts      []int  `json:"attempts"`
	BestIndex     int    `json:"best_index"`
	WorstIndex    int    `json:"worst_index"`
}

func (u *UpdateDiyRankings) apiGetWCAResults(wcaID string) (*PersonBestResults, error) {
	wcaID = strings.ToUpper(wcaID)
	var resp []WCAResults
	if err := utils2.HTTPRequestWithJSON(http.MethodGet, fmt.Sprintf(wcaUrlFormat, wcaID), nil, nil, nil, &resp); err != nil {
		return nil, err
	}

	var out = PersonBestResults{
		PersonName: "",
		Best:       make(map[string]Results),
		Avg:        make(map[string]Results),
	}

	for _, v := range resp {
		out.PersonName = v.Name

		// 无数据的时候
		if _, ok := out.Best[v.EventId]; (!ok && v.Best > 0) || (v.Best > 0 && ok && utils.IsBestResult(v.EventId, v.Best, out.Best[v.EventId].Best)) {
			out.Best[v.EventId] = Results{
				EventId:    v.EventId,
				Best:       v.Best,
				BestStr:    utils.ResultsTimeFormat(v.Best, v.EventId),
				PersonName: v.Name,
				PersonId:   v.WcaId,
			}
		}
		if _, ok := out.Avg[v.EventId]; (!ok && v.Average > 0) || (v.Average > 0 && ok && utils.IsBestResult(v.EventId, v.Average, out.Avg[v.EventId].Average)) {
			out.Avg[v.EventId] = Results{
				EventId:    v.EventId,
				Average:    v.Average,
				AverageStr: utils.ResultsTimeFormat(v.Average, v.EventId),
				PersonName: v.Name,
				PersonId:   v.WcaId,
			}
		}

	}

	return &out, nil
}

func (u *UpdateDiyRankings) apiGetAllResult(WcaIDs []string) map[string]PersonBestResults {
	var out = make(map[string]PersonBestResults)

	WcaIDs = utils2.RemoveRepeatedElement(WcaIDs)

	var resultsCh []*PersonBestResults

	for _, wcaId := range WcaIDs {
		log.Printf("[apiGetAllResult] %+v\n", wcaId)
		res, err := u.apiGetWCAResults(wcaId)
		if err != nil {
			log.Printf("[apiGetAllResult] get wca %s error %+v\n", wcaId, err)
			continue
		}
		resultsCh = append(resultsCh, res)
		time.Sleep(time.Second)
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
		var bests []Results
		var avgs []Results

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
