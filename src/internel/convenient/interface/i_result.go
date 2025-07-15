package _interface

import (
	"fmt"
	"sort"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"

	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
)

type ResultI interface {
	AllPlayerBestResult(results []result.Results, players []user.User) (best PlayerBestResult, all []PlayerBestResult) // 获取所有成绩列表中，对应玩家的所有最佳成绩汇总
	PlayerBestResult(playerId uint, events []string, year *int) (PlayerBestResult, error)                              // 获取玩家最佳成绩
	PlayerNemesisWithID(playerId uint, events []string) (out []Nemesis)
	PlayerNemesis(player PlayerBestResult, all []PlayerBestResult, events map[string]event.Event, com bool) []Nemesis
	KinChSor(best PlayerBestResult, events []event.Event, players []PlayerBestResult) []KinChSorResult
	KinChSorWithPlayer(playerId uint, events []string) (KinChSorResult, error)

	// 以下都是带缓存的
	SelectKinChSor(page int, size int, events []event.Event) ([]KinChSorResult, int)                       // 获取sor
	SelectAllPlayerBestResult() (best PlayerBestResult, all []PlayerBestResult)                            // 获取一个玩家所有成绩
	SelectBestResultsWithEventSort() (best map[EventID][]result.Results, avg map[EventID][]result.Results) // 获取所有成绩的排名
	SelectBestResultsWithEventSortWithPlayer(cubeId string) PlayerBestResult                               // 获取用户最佳成绩
	SelectUserResultDetail(cubeId string, year *int) UserResultDetail                                      // 获取用户详细成绩信息
	SelectCompsResult(id uint) map[EventID]map[int][]result.Results                                        // 比赛成绩 map[项目] map[轮次] 成绩列表
	SelectPlayerEndYears(id uint, year int) (PlayerEndYears, error)                                        // 某一年的年终总结

	SelectAllPlayerBestResultWithGroup(groupId uint) (best PlayerBestResult, all []PlayerBestResult) // 获取一个群组的比赛
}

type ResultIter struct {
	DB    *gorm.DB
	Cache *cache.Cache
}

func (c *ResultIter) AllPlayerBestResult(results []result.Results, players []user.User) (best PlayerBestResult, all []PlayerBestResult) {
	cacheResult := make(map[uint]PlayerBestResult)

	for _, player := range players {
		cacheResult[player.ID] = PlayerBestResult{
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
		if res.DBest() || res.Best == 0 {
			continue
		}
		if single, ok := cacheResult[res.UserID].Single[res.EventID]; !ok || res.IsBest(single) {
			cacheResult[res.UserID].Single[res.EventID] = res
		}

		if res.DAvg() || res.EventRoute.RouteMap().Repeatedly || res.Average == 0 {
			continue
		}
		if avg, ok := cacheResult[res.UserID].Avgs[res.EventID]; !ok || res.IsBestAvg(avg) {
			cacheResult[res.UserID].Avgs[res.EventID] = res
		}
	}

	best = PlayerBestResult{
		Single: make(map[EventID]result.Results),
		Avgs:   make(map[EventID]result.Results),
	}
	for _, player := range players {
		all = append(all, cacheResult[player.ID])

		for e, single := range cacheResult[player.ID].Single {
			if s, ok := best.Single[e]; !ok || single.IsBest(s) {
				best.Single[e] = single
			}
		}
		for e, avg := range cacheResult[player.ID].Avgs {
			if a, ok := best.Avgs[e]; !ok || avg.IsBestAvg(a) {
				best.Avgs[e] = avg
			}
		}
	}

	return
}

func (c *ResultIter) PlayerBestResult(playerId uint, events []string, year *int) (PlayerBestResult, error) {
	var player user.User
	if err := c.DB.Where("id = ?", playerId).First(&player).Error; err != nil {
		return PlayerBestResult{}, err
	}

	var results []result.Results
	db := c.DB.Where("user_id = ?", playerId)
	if len(events) > 0 {
		db = db.Where("event_id in ?", events)
	}

	if year != nil {
		timeLimit := time.Date(*year, 1, 1, 0, 0, 0, 0, time.UTC)
		db = db.Where("created_at < ?", timeLimit)
	}

	db.Find(&results)
	b, _ := c.AllPlayerBestResult(results, []user.User{player})
	return b, nil
}

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

	playerBest, err := c.PlayerBestResult(playerId, events, nil)
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
	var out []KinChSorResult

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
	}

	for _, player := range players {
		k := KinChSorResult{
			Player:  player.Player,
			Result:  0,
			Results: make([]KinChSorResultWithEvent, 0),
		}

		for _, e := range hasResultEvents {
			mp := e.BaseRouteType.RouteMap()
			kr := KinChSorResultWithEvent{Event: e.ID}
			s, ok := player.Single[e.ID]
			if ok {
				if mp.Repeatedly {
					// 多盲分+(60-用时)/60
					bestResult := best.Single[e.ID].Best + ((3600 - best.Single[e.ID].BestRepeatedlyTime) / 3600)
					playerResult := s.Best + ((3600 - s.BestRepeatedlyTime) / 3600)
					kr.Result = (playerResult / bestResult) * 100
				} else if mp.WithBest {
					kr.Result = (best.Single[e.ID].Best / s.Best) * 100
				} else {
					if a, ok2 := player.Avgs[e.ID]; ok2 {
						kr.Result = (best.Avgs[e.ID].Average / a.Average) * 100
					}
				}
				kr.IsBest = kr.Result == 100.0
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

	for i := 0; i < len(out); i++ {
		out[i].Rank = i + 1
	}
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

func (c *ResultIter) SelectKinChSor(page int, size int, events []event.Event) ([]KinChSorResult, int) {
	if len(events) == 0 {
		return nil, 0
	}
	var keys = "SelectKinChSor"
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

	var results []result.Results
	var userIds []uint

	c.DB.Where("event_id in ?", eventIds).Find(&results)
	c.DB.Model(&result.Results{}).Distinct("user_id").Where("event_id in ?", eventIds).Find(&userIds)
	var evs []event.Event
	c.DB.Where("id in ?", eventIds).Find(&evs)
	var players []user.User
	c.DB.Where("id in ?", userIds).Find(&players)

	best, all := c.AllPlayerBestResult(results, players)

	data := c.KinChSor(best, evs, all)

	c.Cache.Set(keys, data, time.Minute*60)
	return utils.Page[KinChSorResult](data, page, size)
}

func (c *ResultIter) SelectAllPlayerBestResult() (best PlayerBestResult, all []PlayerBestResult) {
	value, ok := c.Cache.Get("SelectAllPlayerBestResult")
	if ok {
		data := value.([2]interface{})
		return data[0].(PlayerBestResult), data[1].([]PlayerBestResult)
	}

	var players []user.User
	var results []result.Results

	c.DB.Where("ban = ?", false).Find(&results)
	c.DB.Find(&players)

	best, all = c.AllPlayerBestResult(results, players)
	c.Cache.Set("SelectAllPlayerBestResult", [2]interface{}{best, all}, time.Minute*30)
	return best, all
}

func (c *ResultIter) SelectBestResultsWithEventSort() (best map[EventID][]result.Results, avg map[EventID][]result.Results) {
	value, ok := c.Cache.Get("SelectBestResultsWithEventSort")
	if ok {
		data := value.([2]interface{})
		return data[0].(map[EventID][]result.Results), data[1].(map[EventID][]result.Results)
	}
	best, avg = make(map[EventID][]result.Results), make(map[EventID][]result.Results)

	_, all := c.SelectAllPlayerBestResult()

	for _, pResult := range all {
		for _, s := range pResult.Single {
			best[s.EventID] = append(best[s.EventID], s)
		}
		for _, a := range pResult.Avgs {
			avg[a.EventID] = append(avg[a.EventID], a)
		}
	}

	for key := range best {
		var b = best[key]
		result.SortResultWithBest(b)
		best[key] = b
	}

	for key := range avg {
		var a = avg[key]
		result.SortResultWithAvg(a)
		avg[key] = a
	}

	c.Cache.Set("SelectBestResultsWithEventSort", [2]interface{}{best, avg}, time.Minute*30)
	return best, avg
}

func (c *ResultIter) SelectBestResultsWithEventSortWithPlayer(cubeId string) PlayerBestResult {
	value, ok := c.Cache.Get("SelectBestResultsWithEventSortWithPlayer_" + cubeId)
	if ok {
		data := value.(map[string]PlayerBestResult)
		return data[cubeId]
	}

	var dict = make(map[string]PlayerBestResult)
	setCubeId := func(cubeId string, PersonName string, userId uint) {
		if _, ok := dict[cubeId]; !ok {
			dict[cubeId] = PlayerBestResult{
				Player: Player{
					PlayerId:   userId,
					PlayerName: PersonName,
					CubeId:     cubeId,
				},
				Single: make(map[EventID]result.Results),
				Avgs:   make(map[EventID]result.Results),
			}
		}
	}

	best, avg := c.SelectBestResultsWithEventSort()
	for _, bb := range best {
		for _, b := range bb {
			setCubeId(b.CubeID, b.PersonName, b.UserID)
			dict[b.CubeID].Single[b.EventID] = b
		}
	}
	for _, aa := range avg {
		for _, a := range aa {
			setCubeId(a.CubeID, a.PersonName, a.UserID)
			dict[a.CubeID].Avgs[a.EventID] = a
		}
	}

	c.Cache.Set("SelectBestResultsWithEventSortWithPlayer_"+cubeId, dict, time.Minute*30)
	return dict[cubeId]
}

func (c *ResultIter) SelectUserResultDetail(cubeId string, year *int) UserResultDetail {
	value, ok := c.Cache.Get("SelectUserResultDetail_" + cubeId)
	if ok {
		return value.(UserResultDetail)
	}

	var out UserResultDetail

	var results []result.Results
	db := c.DB.Where("ban = ?", false).Where("cube_id = ?", cubeId)

	if year != nil {
		timeLimit := time.Date(*year, 1, 1, 0, 0, 0, 0, time.UTC)
		db = db.Where("created_at < ?", timeLimit)
	}

	db.Find(&results)

	var compIds = make(map[uint]struct{})

	for _, r := range results {
		if r.EventRoute.RouteMap().Repeatedly {
			continue
		}
		compIds[r.CompetitionID] = struct{}{}
		for _, rr := range r.Result {

			if rr <= result.DNS {
				continue
			}
			out.RestoresNum += 1
			if rr > result.DNF {
				out.SuccessesNum += 1
			}
		}
	}
	out.Matches = len(compIds)

	// todo PodiumNum
	c.Cache.Set("SelectUserResultDetail_", out, time.Minute*15)

	return out
}

func (c *ResultIter) SelectCompsResult(id uint) map[EventID]map[int][]result.Results {
	key := "SelectCompsResult_" + fmt.Sprint(id)
	if value, ok := c.Cache.Get(key); ok {
		return value.(map[EventID]map[int][]result.Results)
	}

	var out = make(map[EventID]map[int][]result.Results)

	var results []result.Results
	c.DB.Where("comp_id = ?", id).Where("ban = ?", false).Find(&results)

	for _, rr := range results {
		if _, ok := out[rr.EventID]; !ok {
			out[rr.EventID] = make(map[int][]result.Results)
		}
		if _, ok := out[rr.EventID][rr.RoundNumber]; !ok {
			out[rr.EventID][rr.RoundNumber] = make([]result.Results, 0)
		}
		out[rr.EventID][rr.RoundNumber] = append(out[rr.EventID][rr.RoundNumber], rr)
	}
	for k, _ := range out {
		for k2, _ := range out[k] {
			result.SortResult(out[k][k2])
		}
	}

	c.Cache.Set(key, out, time.Minute*15)
	return out
}

func (c *ResultIter) SelectPlayerEndYears(id uint, year int) (PlayerEndYears, error) {

	var eventIds []string
	c.DB.Model(&event.Event{}).Distinct("id").Where("is_comp = ?", true).Find(&eventIds)
	out := PlayerEndYears{
		Player:       Player{},
		HasYears:     nil,
		CurYear:      0,
		CompNum:      0,
		RestoresNum:  0,
		SuccessesNum: 0,
		Single:       nil,
		Avg:          nil,
		Ao12:         nil,
		Ao50:         nil,
		Ao100:        nil,
	}

	playerBest, err := c.PlayerBestResult(id, eventIds, &year)
	if err != nil {
		return out, err
	}

	c.SelectUserResultDetail(playerBest.CubeId, &year)

	return out, nil
}

func (c *ResultIter) SelectAllPlayerBestResultWithGroup(groupId uint) (best PlayerBestResult, all []PlayerBestResult) {
	key := "SelectAllPlayerBestResultWithGroup_" + fmt.Sprint(groupId)

	if value, ok := c.Cache.Get(key); ok {
		data := value.([2]interface{})
		return data[0].(PlayerBestResult), data[1].([]PlayerBestResult)
	}

	// 查分组、比赛列表和成绩列表
	var group competition.CompetitionGroup
	if c.DB.Where("id = ?", group).First(&group).Error != nil {
		return
	}
	var comps []competition.Competition
	if c.DB.Where("group_id = ?", groupId).Find(&comps).Error != nil {
		return
	}
	var compIds []uint
	for _, comp := range comps {
		compIds = append(compIds, comp.ID)
	}
	var results []result.Results
	if c.DB.Where("comp_id IN ?", compIds).Where("ban = ?", false).Find(&results).Error != nil {
		return
	}

	var players []user.User
	c.DB.Find(&players)

	best, all = c.AllPlayerBestResult(results, players)
	c.Cache.Set(key, [2]interface{}{best, all}, time.Minute*30)
	return
}
