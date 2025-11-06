package _interface

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/wca"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/internel/wca_api"

	wca_utils "github.com/guojia99/cubing-pro/src/internel/database/wca_model/utils"
)

type WCAResultI interface {
	SelectSeniorKinChSor(page int, size int, age int, events []event.Event) ([]KinChSorResult, int) // 获取sor
	SelectKinchWithWcaIDs(wcaIds []string, page int, size int, events []event.Event) ([]KinChSorResult, int)
}

func (c *ResultIter) SelectSeniorKinChSor(page int, size int, age int, events []event.Event) ([]KinChSorResult, int) {
	if len(events) == 0 {
		return nil, 0
	}
	var keys = fmt.Sprintf("SelectSeniorKinChSor_age%d_", age)
	var eventIds []string
	for _, ev := range events {
		keys += fmt.Sprintf("_%s", ev.ID)
		eventIds = append(eventIds, ev.ID)
	}

	value, ok := c.Cache.Get(keys)
	if ok {
		data := value.([]KinChSorResult)
		return utils.Page[KinChSorResult](data, page, size)
	}

	var evs []event.Event
	c.DB.Where("id in ?", eventIds).Find(&evs)
	var evsMap = make(map[string]event.Event)
	for _, ev := range evs {
		evsMap[ev.ID] = ev
	}

	best, allSeniors, err := wca_api.GetSeniorsWithEventsAndGroup(age, eventIds)
	if err != nil {
		return nil, 0
	}

	// 最佳成绩
	bestResult := PlayerBestResult{
		Single: make(map[EventID]result.Results),
		Avgs:   make(map[EventID]result.Results),
	}
	for ev, si := range best.Single {
		eve := evsMap[ev]
		bestResult.Single[ev] = seniorRankToResult(si, eve)
	}
	for ev, ai := range best.Average {
		eve := evsMap[ev]
		bestResult.Avgs[ev] = seniorRankToResult(ai, eve)
	}

	// 全部成绩预处理
	var allPlayer = make(map[string]PlayerBestResult)

	for ev, as := range allSeniors {
		eve := evsMap[ev]
		for _, a := range as {
			if _, ageOk := a.Single[age]; !ageOk {
				continue
			}

			if _, ok1 := allPlayer[a.Id]; !ok1 {
				allPlayer[a.Id] = PlayerBestResult{
					Player: Player{
						WcaID:      a.Id,
						WcaName:    a.Name,
						PlayerName: a.Name,
					},
					Single: make(map[EventID]result.Results),
					Avgs:   make(map[EventID]result.Results),
				}
			}

			if single, sOk := a.Single[age][ev]; sOk {
				allPlayer[a.Id].Single[ev] = seniorRankToResult(single, eve)
			}

			if _, ageOk := a.Average[age]; !ageOk {
				continue
			}

			if avg, aOk := a.Average[age][ev]; aOk {
				allPlayer[a.Id].Avgs[ev] = seniorRankToResult(avg, eve)
			}
		}
	}

	// 处理数据
	var all []PlayerBestResult
	for _, a := range allPlayer {
		all = append(all, a)
	}
	data := c.KinChSor(&bestResult, evs, all)

	c.Cache.Set(keys, data, time.Minute*60)
	return utils.Page[KinChSorResult](data, page, size)

}

func seniorRankToResult(rank wca_api.SeniorRank, ev event.Event) result.Results {
	var out = result.Results{
		Best:                    result.DNF,
		Average:                 result.DNF,
		BestRepeatedlyReduction: 0,
		BestRepeatedlyTry:       0,
		BestRepeatedlyTime:      0,
		EventID:                 ev.ID,
		EventName:               ev.Name,
		EventRoute:              ev.BaseRouteType,
	}

	switch rank.Type {
	case "single":
		if ev.BaseRouteType.RouteMap().Repeatedly {
			// like 21/23 in 53:12
			cut := strings.Split(rank.Best, "in")
			if len(cut) != 2 {
				return out
			}
			out.BestRepeatedlyTime = result.TimeParserS2F(cut[1])
			sps := strings.Split(cut[0], "/")
			a, b := strings.ReplaceAll(sps[0], " ", ""), strings.ReplaceAll(sps[1], " ", "")
			out.BestRepeatedlyReduction, _ = strconv.ParseFloat(a, 64)
			out.BestRepeatedlyTry, _ = strconv.ParseFloat(b, 64)

			out.Best = out.BestRepeatedlyReduction - (out.BestRepeatedlyTry - out.BestRepeatedlyReduction)
		} else {
			out.Best = result.TimeParserS2F(rank.Best)
		}
	case "average":
		out.Average = result.TimeParserS2F(rank.Best)
	}

	return out
}

func (c *ResultIter) SelectKinchWithWcaIDs(wcaIds []string, page int, size int, events []event.Event) ([]KinChSorResult, int) {

	key, err := utils.MakeCacheKey(wcaIds, events)
	if err == nil && key != "" {
		value, ok := c.Cache.Get(key)
		if ok {
			data := value.([]KinChSorResult)
			return utils.Page[KinChSorResult](data, page, size)
		}
	}

	var dbWcaResults []wca.WCAResult
	if err = c.DB.Where("wca_id in ?", wcaIds).Find(&dbWcaResults).Error; err != nil {
		return nil, 0
	}
	var eventMap = make(map[string]event.Event)
	for _, ev := range events {
		eventMap[ev.ID] = ev
	}

	var all []PlayerBestResult
	for _, a := range dbWcaResults {
		all = append(all, wcaResultToPlayerBestResult(a, eventMap))
	}
	data := c.KinChSor(nil, events, all)

	c.Cache.Set(key, data, time.Minute*60)
	return utils.Page[KinChSorResult](data, page, size)
}

func wcaResultToPlayerBestResult(wcaResult wca.WCAResult, eventMap map[string]event.Event) PlayerBestResult {
	out := PlayerBestResult{
		Player: Player{
			WcaID:      wcaResult.WcaID,
			WcaName:    wcaResult.PersonBestResults.PersonName,
			PlayerName: wcaResult.PersonBestResults.PersonName,
		},
		Single: make(map[EventID]result.Results),
		Avgs:   make(map[EventID]result.Results),
	}

	for ev, best := range wcaResult.PersonBestResults.Best {
		var newBest = result.Results{
			Best:                    result.DNF,
			Average:                 result.DNF,
			BestRepeatedlyReduction: 0,
			BestRepeatedlyTry:       0,
			BestRepeatedlyTime:      0,
			EventID:                 eventMap[ev].ID,
			EventName:               eventMap[ev].Name,
			EventRoute:              eventMap[ev].BaseRouteType,
		}

		avg, hasAvg := wcaResult.PersonBestResults.Avg[ev]

		switch ev {
		case "333mbf":
			solved, attempted, seconds, _ := wca_utils.Get333MBFResult(best.Best)
			newBest.BestRepeatedlyReduction = float64(solved)
			newBest.BestRepeatedlyTry = float64(attempted)
			newBest.BestRepeatedlyTime = float64(seconds)
			newBest.Best = newBest.BestRepeatedlyReduction - (newBest.BestRepeatedlyTry - newBest.BestRepeatedlyReduction)
		default:
			newBest.Best = float64(best.Best) / 100.0
			if hasAvg {
				newBest.Average = float64(avg.Average) / 100.0
			}
		}

		out.Single[ev] = newBest
		if hasAvg {
			out.Avgs[ev] = newBest
		}
	}

	return out
}
