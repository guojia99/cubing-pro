package _interface

import (
	"fmt"
	"gorm.io/gorm"
	"sort"

	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
)

type ResultI interface {
	AllPlayerBestResult(results []result.Results, players []user.User) (best PlayerBestResult, all []PlayerBestResult) // 获取所有成绩列表中，对应玩家的所有最佳成绩汇总
	PlayerBestResult(playerId uint, events []string) (PlayerBestResult, error)                                         // 获取玩家最佳成绩
	PlayerNemesisWithID(playerId uint, events []string) (out []Nemesis)
	PlayerNemesis(player PlayerBestResult, all []PlayerBestResult, events map[string]event.Event, com bool) []Nemesis
	KinChSor(best PlayerBestResult, events []event.Event, players []PlayerBestResult) []KinChSorResult
	KinChSorWithPlayer(playerId uint, events []string) (KinChSorResult, error)
}

type ResultIter struct {
	DB *gorm.DB
}

func (c *ResultIter) AllPlayerBestResult(results []result.Results, players []user.User) (best PlayerBestResult, all []PlayerBestResult) {
	cache := make(map[uint]PlayerBestResult)

	for _, player := range players {
		cache[player.ID] = PlayerBestResult{
			Player: Player{
				PlayerId:   player.ID,
				CubeId:     player.CubeID,
				PlayerName: player.Name,
			},
			Single: make(map[EventID]result.Results),
			Avgs:   make(map[EventID]result.Results),
		}
	}

	for _, res := range results {
		if res.DBest() {
			continue
		}
		if single, ok := cache[res.UserID].Single[res.EventID]; !ok || res.IsBest(single) {
			cache[res.UserID].Single[res.EventID] = res
		}

		if res.DAvg() {
			continue
		}
		if avg, ok := cache[res.UserID].Avgs[res.EventID]; !ok || res.IsBestAvg(avg) {
			cache[res.UserID].Avgs[res.EventID] = res
		}
	}

	best = PlayerBestResult{
		Single: make(map[EventID]result.Results),
		Avgs:   make(map[EventID]result.Results),
	}
	for _, player := range players {
		all = append(all, cache[player.ID])

		for e, single := range cache[player.ID].Single {
			if s, ok := best.Single[e]; !ok || single.IsBest(s) {
				best.Single[e] = single
			}
		}
		for e, avg := range cache[player.ID].Avgs {
			if a, ok := best.Avgs[e]; !ok || avg.IsBestAvg(a) {
				best.Avgs[e] = avg
			}
		}
	}
	return
}

func (c *ResultIter) PlayerBestResult(playerId uint, events []string) (PlayerBestResult, error) {
	var player user.User
	if err := c.DB.Where("id = ?", playerId).First(&player).Error; err != nil {
		return PlayerBestResult{}, err
	}

	var results []result.Results
	c.DB.Where("user_id = ?", playerId).Where("event_id in ?", events).Find(&results)
	b, _ := c.AllPlayerBestResult(results, []user.User{player})
	return b, nil
}

// PlayerNemesis onlyCom 仅比较都有成绩的项目
func (c *ResultIter) PlayerNemesis(player PlayerBestResult, all []PlayerBestResult, events map[string]event.Event, com bool) (out []Nemesis) {
	for _, pr := range all {
		if player.PlayerId == pr.PlayerId {
			continue
		}

		win, has := false, false // has代表至少有一个项目做比较
		for _, ev := range events {
			playerSingle, ok1 := player.Single[ev.ID]
			prSingle, ok2 := pr.Single[ev.ID]
			// 双方必须都有该项目，否则无效
			if com && (!ok1 || !ok2) {
				continue
			}
			// 双方都没有
			if !ok1 && !ok2 {
				continue
			}

			has = true
			// 如果该玩家没有该项目，直接算输
			if !ok1 {
				continue
			}
			// 如果宿敌没有该项目，则失去宿敌身份
			if !ok2 {
				win = true
				break
			}

			// 计算单次项目
			mp := ev.BaseRouteType.RouteMap()

			// 其他项目的单次
			if playerSingle.IsBest(prSingle) {
				win = true
			}
			if mp.Repeatedly {
				continue
			}

			// 其他项目的平均
			playerAvg, ok3 := player.Avgs[ev.ID]
			prAvg, ok4 := pr.Avgs[ev.ID]
			if !ok3 {
				continue
			}
			if !ok4 || playerAvg.IsBestAvg(prAvg) {
				win = true
				break
			}
		}

		if win || !has {
			continue
		}
		out = append(
			out, Nemesis{
				PlayerBestResult: pr,
			},
		)
	}

	return
}

func (c *ResultIter) PlayerNemesisWithID(playerId uint, events []string) (out []Nemesis) {

	playerBest, err := c.PlayerBestResult(playerId, events)
	if err != nil {
		return
	}

	var results []result.Results
	var userIds []uint
	c.DB.Where("event_id in ?", events).Find(&results)
	c.DB.Model(&result.Results{}).Distinct("user_id").Where("event_id in ?", events).Find(&userIds)

	var evs []event.Event
	c.DB.Where("id in ?", events).Find(&evs)
	var eventMap = make(map[string]event.Event)
	for _, e := range evs {
		eventMap[e.ID] = e
	}

	var players []user.User
	c.DB.Where("id in ?", userIds).Find(&players)

	_, all := c.AllPlayerBestResult(results, players)

	out = c.PlayerNemesis(playerBest, all, eventMap, false)
	return
}

func (c *ResultIter) KinChSor(best PlayerBestResult, events []event.Event, players []PlayerBestResult) []KinChSorResult {
	var out = make([]KinChSorResult, 0)

	// 过滤有有效成绩的成绩
	var hasResultEvents []event.Event
	for _, e := range events {
		mp := e.BaseRouteType.RouteMap()
		// 单次最佳项目
		if mp.WithBest {
			if _, ok := best.Single[e.ID]; ok {
				hasResultEvents = append(hasResultEvents, e)
				continue
			}
		}
		if _, ok := best.Avgs[e.ID]; ok {
			hasResultEvents = append(hasResultEvents, e)
			continue
		}
		fmt.Printf("not %s\n", e.ID)
	}

	for _, player := range players {
		k := KinChSorResult{
			Player:  player.Player,
			Result:  0,
			Results: make([]KinChSorResultWithEvent, 0),
		}

		for _, e := range hasResultEvents {
			mp := e.BaseRouteType.RouteMap()
			kr := KinChSorResultWithEvent{Event: e}
			s, ok := player.Single[e.ID]
			if ok {
				if mp.Repeatedly {
					// 多盲分+(60-用时)/60
					bestResult := best.Single[e.ID].Best + ((3600 - best.Single[e.ID].BestRepeatedlyTime) / 3600)
					playerResult := s.Best + ((3600 - s.BestRepeatedlyTime) / 3600)
					kr.Result = (playerResult / bestResult) * 100

					kr.IsBest = kr.Result == 100.0
				} else if mp.WithBest {
					kr.Result = (best.Single[e.ID].Best / s.Best) * 100
				} else {
					a, ok := player.Avgs[e.ID]
					if ok {
						kr.Result = (best.Avgs[e.ID].Average / a.Average) * 100
					}
				}
			}

			k.Results = append(k.Results, kr)
		}

		for _, kr := range k.Results {
			k.Result += kr.Result / float64(len(k.Results))
		}
		out = append(out, k)
	}

	sort.Slice(
		out, func(i, j int) bool {
			return out[i].Result > out[j].Result
		},
	)
	return out
}

func (c *ResultIter) KinChSorWithPlayer(playerId uint, events []string) (KinChSorResult, error) {
	var results []result.Results
	var userIds []uint

	c.DB.Where("event_id in ?", events).Find(&results)
	c.DB.Model(&result.Results{}).Distinct("user_id").Where("event_id in ?", events).Find(&userIds)
	var evs []event.Event
	c.DB.Where("id in ?", events).Find(&evs)
	var players []user.User
	c.DB.Where("id in ?", userIds).Find(&players)

	best, all := c.AllPlayerBestResult(results, players)

	sor := c.KinChSor(best, evs, all)

	for _, s := range sor {
		if s.PlayerId == playerId {
			return s, nil
		}
	}
	return KinChSorResult{}, fmt.Errorf("not found sor")
}
