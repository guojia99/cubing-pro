package pktimer

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	pktimerDB "github.com/guojia99/cubing-pro/src/internel/database/model/pktimer"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	utils2 "github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/log"
	"github.com/guojia99/cubing-pro/src/robot/types"
)

func (p *PkTimer) checkInPkTimer(msg types.InMessage) bool {
	return p.getMessageDBPkTimer(msg) != nil
}

func (p *PkTimer) initPkTimer(msg types.InMessage) error {
	var pkTimerResult pktimerDB.PkTimerResult
	if p.Svc.DB.Where("group_id = ?", msg.GroupIDStr()).Where("running = ?", true).First(&pkTimerResult).Error == nil {
		if time.Since(pkTimerResult.LastRunning) < time.Minute*20 {
			p.sendMessage(msg.NewOutMessage(fmt.Sprintf("请结束上一次pk后再开始pk")))
			return nil
		}
		pkTimerResult.Running = false
		p.Svc.DB.Save(&pkTimerResult)
	}

	usr, err := p.getMsgUser(msg)
	if err != nil {
		p.sendMessage(msg.NewOutMessagef("用户`%d | %s`未绑定，请绑定Cubing Pro后再使用", msg.QQ, msg.QQBot))
		return err
	}

	m := utils2.ReplaceAll(msg.Message, "", key)
	slice := utils2.Split(m, " ")
	var count int
	var ev string
	if len(slice) == 0 || len(slice) >= 3 {
		p.sendMessage(msg.NewOutMessage("请输入: `pktimer 333 5`来代表pk5次3速"))
		return nil
	}
	ev = slice[0]
	if len(slice) == 2 {
		count, _ = strconv.Atoi(slice[1])
	}
	if count <= 0 {
		count = 5
	}
	if count > 20 {
		count = 20
	}

	var eve event.Event
	if err = p.Svc.DB.Where("id = ?", ev).Or("name = ?", ev).First(&eve).Error; err != nil {
		p.sendMessage(msg.NewOutMessagef("未找到项目 %s", ev))
		return nil
	}
	if eve.BaseRouteType.RouteMap().Repeatedly || !eve.IsWCA {
		p.sendMessage(msg.NewOutMessage("不支持的项目"))
		return nil
	}

	player := pktimerDB.Player{
		QQ:       msg.QQ,
		QQBot:    msg.QQBot,
		UserName: usr.Name,
		UserId:   usr.ID,
		Results:  make([]float64, 0),
	}
	newPKTimerResult := pktimerDB.PkTimerResult{
		GroupID:     msg.GroupIDStr(),
		Running:     true,
		LastRunning: time.Now(),
		StartPerson: usr.Name,
		PkResults: pktimerDB.PkResults{
			Players:      make([]pktimerDB.Player, 0),
			Event:        eve,
			Count:        count,
			CurCount:     0,
			FirstMessage: msg,
		},
	}
	newPKTimerResult.PkResults.Players = append(newPKTimerResult.PkResults.Players, player)
	p.Svc.DB.Save(&newPKTimerResult)

	p.sendMessage(msg.NewOutMessage(getIniterMessage(&newPKTimerResult)))
	return nil
}

func (p *PkTimer) in(msg types.InMessage) error {
	pkTimerResult := p.getMessageDBPkTimer(msg)

	if pkTimerResult.Start {
		p.sendMessage(msg.NewOutMessage("本次pk已经开始，请下一次pk及时加入"))
		return nil
	}
	usr, err := p.getMsgUser(msg)
	if err != nil {
		p.sendMessage(msg.NewOutMessagef("用户`%d | %s`未绑定，请绑定Cubing Pro后再使用", msg.QQ, msg.QQBot))
		return err
	}

	player := pktimerDB.Player{
		QQ:       msg.QQ,
		QQBot:    msg.QQBot,
		UserName: usr.Name,
		UserId:   usr.ID,
	}

	for _, pl := range pkTimerResult.PkResults.Players {
		if pl.UserId == usr.ID {
			p.sendMessage(msg.NewOutMessage("你已经加入本次pk"))
			p.sendMessage(msg.NewOutMessage(getIniterMessage(pkTimerResult)))
			return nil
		}
	}

	pkTimerResult.PkResults.Players = append(pkTimerResult.PkResults.Players, player)
	pkTimerResult.LastRunning = time.Now()
	p.Svc.DB.Save(&pkTimerResult)
	p.sendMessage(msg.NewOutMessage(getIniterMessage(pkTimerResult)))
	return nil
}

func (p *PkTimer) startPkTimer(msg types.InMessage) error {
	pkTimerResult := p.getMessageDBPkTimer(msg)
	if pkTimerResult.Start {
		return nil
	}
	pkTimerResult.Start = true
	pkTimerResult.LastRunning = time.Now()
	p.sendMessage(msg.NewOutMessage("本次pk已开始!!"))
	p.Svc.DB.Save(&pkTimerResult)
	p.sendScrambleMessage(msg)
	return nil
}

// endPKTimerMessage  结算
func (p *PkTimer) endPKTimerMessage(res *pktimerDB.PkTimerResult) string {
	out := fmt.Sprintf("PK结果: %s(%d/%d)\n", res.PkResults.Event.Cn, res.PkResults.CurCount, res.PkResults.Count)

	// 计算每个人的成绩
	// 跳过一把的不会有平均
	rm := res.PkResults.Event.BaseRouteType.RouteMap()
	for idx, pl := range res.PkResults.Players {
		best, avg := result.GetBestAndAvg(pl.Results, rm)

		res.PkResults.Players[idx].Best = best
		res.PkResults.Players[idx].Average = avg
	}

	sort.Slice(res.PkResults.Players, func(i, j int) bool {
		p1 := res.PkResults.Players[i]
		p2 := res.PkResults.Players[j]

		if p1.Best <= result.DNF {
			return false
		}
		if p2.Best <= result.DNF {
			return true
		}
		if rm.WithBest {
			return p1.Best < p2.Best
		}

		if p1.Average <= result.DNF {
			return false
		}
		if p2.Average <= result.DNF {
			return true
		}
		if p1.Average == p2.Average {
			return p1.Best < p2.Best
		}
		return p1.Average < p2.Average
	})

	for idx, pl := range res.PkResults.Players {
		out += fmt.Sprintf("%d. %s 成绩 (%s / %s)\n", idx+1, pl.UserName, result.TimeParserF2S(pl.Best), result.TimeParserF2S(pl.Average))
	}
	return out
}

func (p *PkTimer) endPkTimer(msg types.InMessage) error {
	pkTimerResult := p.getMessageDBPkTimer(msg)
	pkTimerResult.Running = false
	p.sendMessage(msg.NewOutMessage(p.endPKTimerMessage(pkTimerResult)))
	p.Svc.DB.Save(&pkTimerResult)
	_ = p.sendPackerMessage(msg, false)
	return nil
}

func (p *PkTimer) sendScrambleMessage(msg types.InMessage) {
	pkTimerResult := p.getMessageDBPkTimer(msg)

	pkTimerResult.PkResults.CurCount += 1
	sc, err := p.Svc.Scramble.ScrambleWithEvent(pkTimerResult.PkResults.Event, 1)
	if err != nil || len(sc) == 0 {
		log.Error(err)
		return
	}
	pkTimerResult.LastRunning = time.Now()
	p.Svc.DB.Save(&pkTimerResult)

	scrMsg := fmt.Sprintf("本次pk第%d / %d把:\n %s", pkTimerResult.PkResults.CurCount, pkTimerResult.PkResults.Count, sc[0])

	img, err := p.Svc.Scramble.Image(sc[0], pkTimerResult.PkResults.Event.ID)
	if err != nil {
		p.sendMessage(msg.NewOutMessage(scrMsg))
		return
	}

	filePath := path.Join(os.TempDir(), fmt.Sprintf("%d.jpg", time.Now().UnixNano()))
	_ = os.WriteFile(filePath, []byte(img), 0644)
	p.sendMessage(msg.NewOutMessageWithImage(scrMsg, filePath))
}

func (p *PkTimer) next(msg types.InMessage) error {
	pkTimerResult := p.getMessageDBPkTimer(msg)

	// 结束了
	if pkTimerResult.PkResults.Count == pkTimerResult.PkResults.CurCount {
		return p.endPkTimer(msg)
	}
	_ = p.sendPackerMessage(msg, true)
	p.sendScrambleMessage(msg)
	return nil
}

func (p *PkTimer) addResult(msg types.InMessage) error {
	pkTimerResult := p.getMessageDBPkTimer(msg)
	if !pkTimerResult.Start {
		return nil
	}
	var can bool
	for _, pl := range pkTimerResult.PkResults.Players {
		if pl.QQBot == msg.QQBot || pl.QQ == msg.QQ {
			can = true
			break
		}
	}
	if !can {
		return nil
	}
	usr, err := p.getMsgUser(msg)
	if err != nil {
		return err
	}

	for idx, pl := range pkTimerResult.PkResults.Players {
		if pl.UserId == usr.ID {
			res := result.TimeParserS2F(msg.Message)
			if res == result.UNT {
				return nil
			}
			if len(pkTimerResult.PkResults.Players[idx].Results) == pkTimerResult.PkResults.CurCount {
				pkTimerResult.PkResults.Players[idx].Results[pkTimerResult.PkResults.CurCount-1] = res
				p.sendMessage(msg.NewOutMessagef("你的第 %d 把 PK 成绩更新成功: %s | (%d/%d)", len(pkTimerResult.PkResults.Players[idx].Results), result.TimeParserF2S(res), len(pkTimerResult.PkResults.Players[idx].Results), pkTimerResult.PkResults.CurCount))
				break
			}
			pkTimerResult.PkResults.Players[idx].Results = append(pkTimerResult.PkResults.Players[idx].Results, res)
			p.sendMessage(msg.NewOutMessagef("你的第 %d 把 PK 成绩记录成功: %s | (%d/%d)", len(pkTimerResult.PkResults.Players[idx].Results), result.TimeParserF2S(res), len(pkTimerResult.PkResults.Players[idx].Results), pkTimerResult.PkResults.CurCount))
			break
		}
	}
	p.Svc.DB.Save(&pkTimerResult)

	// 判断是否本轮所有玩家已经录入完成
	var hasResult int
	for _, pl := range pkTimerResult.PkResults.Players {
		if len(pl.Results) == pkTimerResult.PkResults.CurCount {
			hasResult += 1
		}
	}
	if hasResult == len(pkTimerResult.PkResults.Players) {
		return p.next(msg)
	}

	return nil
}

// 发货简报
func (p *PkTimer) sendPackerMessage(msg types.InMessage, inCur bool) error {
	pkTimerResult := p.getMessageDBPkTimer(msg)
	var out []string
	if inCur {
		out = getCurPackerMessage(pkTimerResult, "cur")
	} else {
		out = getAllPackerMessage(pkTimerResult)
	}
	for _, o := range out {
		p.sendMessage(msg.NewOutMessage(o))
	}
	return nil
}

// pktimer指令
// 1. 开启pktimer模式， 群聊将不再接受pktimer的之外的指令 pktimer 333 12 pktimer {项目} 把数
// 2. 群员可以加入参与 指令： “加入”
// 3. 开启者可以发送“开始”

func (p *PkTimer) runMessage(msg types.InMessage) error {
	switch msg.Message {
	case start:
		return p.startPkTimer(msg)
	case end:
		return p.endPkTimer(msg)
	case add:
		return p.in(msg)
	case next:
		return p.next(msg)
	default:
		return p.addResult(msg)
	}
}
