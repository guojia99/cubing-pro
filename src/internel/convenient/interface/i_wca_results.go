package _interface

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	wca_model "github.com/guojia99/cubing-pro/src/internel/database/model/wca"
	wca_utils "github.com/guojia99/cubing-pro/src/internel/database/model/wca/utils"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/internel/wca_api"
)

type SelectSorWithWcaIDsOption struct {
	Events     []string //
	WithSingle bool     // 使用单次成绩计算
	WithAvg    bool     // 使用平均成绩
}

type WCAResultI interface {
	SelectSeniorKinChSor(page int, size int, age int, events []event.Event) ([]KinChSorResult, int) // 获取sor
	SelectKinchWithWcaIDs(wcaIds []string, page int, size int, events []event.Event) ([]KinChSorResult, int)

	// SelectSorWithWcaIDs with nr, asr, wr
	SelectSorWithWcaIDs(wcaIds []string, page int, size int, opt SelectSorWithWcaIDsOption) ([]SorResult, int)
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

	var dbWcaResults []wca_model.WCAResult
	for idx := range wcaIds {
		wcaIds[idx] = strings.ToUpper(wcaIds[idx])
	}

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

func wcaResultToPlayerBestResult(wcaResult wca_model.WCAResult, eventMap map[string]event.Event) PlayerBestResult {
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
		case "333fm":
			newBest.Best = float64(best.Best)
			if hasAvg {
				newBest.Average = float64(avg.Average) / 100.0
			}
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

func (c *ResultIter) SelectSorWithWcaIDs(wcaIds []string, page int, size int, opt SelectSorWithWcaIDsOption) ([]SorResult, int) {
	if (!opt.WithSingle && !opt.WithAvg) || len(opt.Events) == 0 {
		return nil, 0
	}

	key, err := utils.MakeCacheKey(wcaIds, opt)
	if err == nil && key != "" {
		value, ok := c.Cache.Get(key)
		if ok {
			data := value.([]SorResult)
			return utils.Page[SorResult](data, page, size)
		}
	}

	var dbWcaResults []wca_model.WCAResult
	if err = c.DB.Where("wca_id in ?", wcaIds).Find(&dbWcaResults).Error; err != nil {
		return nil, 0
	}

	// 统计每一个项目的排名并排序
	var singleEvents = make(map[string][]wca_model.Results)
	var avgEvents = make(map[string][]wca_model.Results)
	for _, ev := range opt.Events {
		var s = make([]wca_model.Results, 0)
		var a = make([]wca_model.Results, 0)

		// 单次
		for _, player := range dbWcaResults {
			if res, ok := player.PersonBestResults.Best[ev]; ok {
				s = append(s, res)
			}
			if res, ok := player.PersonBestResults.Avg[ev]; ok {
				a = append(a, res)
			}
		}
		RankByValue(s, func(md wca_model.Results) int { return md.WorldRank }, func(m *wca_model.Results, i int) { m.Rank = i }, false)
		RankByValue(a, func(md wca_model.Results) int { return md.WorldRank }, func(m *wca_model.Results, i int) { m.Rank = i }, false)

		if len(s) > 0 {
			singleEvents[ev] = s
		}
		if len(a) > 0 {
			avgEvents[ev] = a
		}
	}

	// 给每个人设置分数
	var data []SorResult
	for _, dbPlayer := range dbWcaResults {
		var sorResult = SorResult{
			Player: Player{
				WcaID:      dbPlayer.WcaID,
				WcaName:    dbPlayer.PersonBestResults.PersonName,
				PlayerName: dbPlayer.PersonBestResults.PersonName,
			},
			Results: make([]SorResultWithEvent, 0),
		}

		for _, ev := range opt.Events {
			if opt.WithSingle && len(singleEvents[ev]) != 0 {
				var sorResultWithEvent = SorResultWithEvent{
					Event:        ev,
					IsBest:       true,
					ResultString: "", // 默认无成绩
					Rank:         len(singleEvents[ev]) + 1,
				}
				for _, res := range singleEvents[ev] {
					if res.PersonId == dbPlayer.WcaID {
						sorResultWithEvent.Rank = res.Rank
						sorResultWithEvent.ResultString = res.BestStr
						continue
					}
				}
				sorResult.Sor += sorResultWithEvent.Rank
				sorResult.Results = append(sorResult.Results, sorResultWithEvent)
			}

			if opt.WithAvg && len(avgEvents[ev]) != 0 {
				var sorResultWithEvent = SorResultWithEvent{
					Event:        ev,
					IsBest:       false,
					ResultString: "", // 默认无成绩
					Rank:         len(avgEvents[ev]) + 1,
				}
				for _, res := range avgEvents[ev] {
					if res.PersonId == dbPlayer.WcaID {
						sorResultWithEvent.Rank = res.Rank
						sorResultWithEvent.ResultString = res.AverageStr
						continue
					}
				}
				sorResult.Sor += sorResultWithEvent.Rank
				sorResult.Results = append(sorResult.Results, sorResultWithEvent)
			}
		}

		data = append(data, sorResult)
	}

	// 排序输出
	RankByValue(data, func(s SorResult) int { return s.Sor }, func(s *SorResult, i int) { s.Rank = i }, false)

	c.Cache.Set(key, data, time.Minute*60)
	return utils.Page[SorResult](data, page, size)
}
