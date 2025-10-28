package _interface

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/internel/wca_api"
)

type WCAResultI interface {
	SelectSeniorKinChSor(page int, size int, age int, events []event.Event) ([]KinChSorResult, int) // 获取sor
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
	data := c.KinChSor(bestResult, evs, all)

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
