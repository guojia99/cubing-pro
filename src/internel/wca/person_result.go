package wca

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/guojia99/cubing-pro/src/internel/database/wca_model/utils"
	utils2 "github.com/guojia99/cubing-pro/src/internel/utils"
)

type (
	Results struct {
		EventId    string `json:"eventId"`
		Best       int    `json:"best"`
		BestStr    string `json:"bestStr"`
		Average    int    `json:"average"`
		AverageStr string `json:"averageStr"`
		PersonName string `json:"personName"`
		PersonId   string `json:"personId"`
	}

	PersonBestResults struct {
		PersonName string             `json:"PersonName"`
		Best       map[string]Results `json:"Best"`
		Avg        map[string]Results `json:"Avg"`
	}

	WCAResults struct {
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
)

const wcaResultUrlFormat = "https://www.worldcubeassociation.org/api/v0/persons/%s/results" // 2017XUYO01

func ApiGetWCAResults(wcaID string) (*PersonBestResults, error) {
	wcaID = strings.ToUpper(wcaID)
	var resp []WCAResults
	if err := utils2.HTTPRequestWithJSON(http.MethodGet, fmt.Sprintf(wcaResultUrlFormat, wcaID), nil, nil, nil, &resp); err != nil {
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
