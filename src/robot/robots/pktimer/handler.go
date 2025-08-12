package pktimer

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
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

	// 	pk := p.getMessageDBPkTimer(msg)
	//	if pk == nil {
	//		return false
	//	}
	//
	//	// 还没开始就全部不允许
	//	if !pk.Start {
	//		return true
	//	}
	//
	//	// 开始了，就看这个人是否在PK
	//	for _, pl := range pk.PkResults.Players {
	//		if pl.QQBot == msg.QQBot {
	//			return true
	//		}
	//	}
	//	return true
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

	if strings.Contains(msg.Message, reload) {
		var lastRunning pktimerDB.PkTimerResult
		if err := p.Svc.DB.Order("created_at DESC").Where("group_id = ?", msg.GroupIDStr()).First(&lastRunning).Error; err != nil {
			p.sendMessage(msg.NewOutMessage(fmt.Sprintf("上一次pk不存在, 请创建后再进行")))
			return nil
		}

		var newPK = pktimerDB.PkTimerResult{
			GroupID:     lastRunning.GroupID,
			Running:     true,
			Start:       false,
			LastRunning: time.Now(),
			StartPerson: lastRunning.StartPerson,
			PkResults: pktimerDB.PkResults{
				Players:      lastRunning.PkResults.Players,
				Event:        lastRunning.PkResults.Event,
				Count:        lastRunning.PkResults.Count,
				CurCount:     0,
				FirstMessage: msg,
			},
			Eps: lastRunning.Eps,
		}

		for idx, _ := range newPK.PkResults.Players {
			newPK.PkResults.Players[idx].Results = make([]float64, 0)
			newPK.PkResults.Players[idx].Average = result.DNF
			newPK.PkResults.Players[idx].Best = result.DNF
		}

		p.Svc.DB.Save(&newPK)
		p.sendMessage(msg.NewOutMessage(getIniterMessage(&newPK)))
		return nil
	}
	return p.createNewPkTimer(msg)
}

func (p *PkTimer) createNewPkTimer(msg types.InMessage) error {

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

	// 找项目
	var events []event.Event
	p.Svc.DB.Find(&events)
	var eve event.Event
	for _, e := range events {
		if !e.IsWCA {
			continue
		}
		if e.ID == ev || e.Name == ev || e.Cn == ev {
			eve = e
			break
		}
		slices := utils2.Split(e.OtherNames, ";")
		for _, v := range slices {
			if v == ev {
				eve = e
				break
			}
		}
		if eve.ID != "" {
			break
		}
	}
	if eve.BaseRouteType.RouteMap().Repeatedly || !eve.IsWCA {
		p.sendMessage(msg.NewOutMessage("不支持的项目"))
		return nil
	}

	// 初始化玩家
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
		Eps: 0.05,
	}
	newPKTimerResult.PkResults.Players = append(newPKTimerResult.PkResults.Players, player)

	// 初始化精度
	var epsMap = map[string]float64{
		"444":   0.03,
		"555":   0.03,
		"minx":  0.03,
		"666":   0.015,
		"777":   0.015,
		"444bf": 0.03,
		"555bf": 0.03,
	}
	if o, ok := epsMap[ev]; ok {
		newPKTimerResult.Eps = o
	}

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
		p.sendMessage(msg.NewOutMessage("本次pk已开始!!"))
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
	rm := pkTimerResult.PkResults.Event.BaseRouteType.RouteMap()
	for idx, pl := range pkTimerResult.PkResults.Players {
		best, avg := result.GetBestAndAvg(pl.Results, rm)

		pkTimerResult.PkResults.Players[idx].Best = best
		pkTimerResult.PkResults.Players[idx].Average = avg
	}

	p.sendMessage(msg.NewOutMessage(p.endPKTimerMessage(pkTimerResult)))
	err := p.Svc.DB.Save(&pkTimerResult).Error
	fmt.Println("end pk -> ", err)

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

	// 处理没有成绩的
	for _, pl := range pkTimerResult.PkResults.Players {
		if len(pl.Results) != pkTimerResult.PkResults.CurCount {
			pl.Results = append(pl.Results, result.DNF)
			p.sendMessage(msg.NewOutMessagef("将进行下一把, %s 本把成绩记录为DNS", pl.UserName))
		}
	}

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
		if pl.Exit {
			hasResult += 1
			continue
		}
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

func (p *PkTimer) updateResult(msg types.InMessage) error {
	message := utils2.ReplaceAll(msg.Message, "", update)
	sl := utils2.Split(message, " ")

	if len(sl) != 2 {
		p.sendMessage(msg.NewOutMessage("修改成绩格式为: `修改 1 1.23`, 代表第一把修改成绩为1.23"))
		return nil
	}

	num, _ := strconv.Atoi(sl[0])
	if num <= 0 {
		p.sendMessage(msg.NewOutMessage("修改成绩格式为: `修改 1 1.23`, 代表第一把修改成绩为1.23"))
		return nil
	}

	usr, err := p.getMsgUser(msg)
	if err != nil {
		return err
	}

	pkTimerResult := p.getMessageDBPkTimer(msg)

	if num > pkTimerResult.PkResults.CurCount {
		p.sendMessage(msg.NewOutMessage("该轮次还未开始，无法修改"))
		return nil
	}

	for idx, pl := range pkTimerResult.PkResults.Players {
		if pl.UserId == usr.ID {
			if pl.Exit {
				p.sendMessage(msg.NewOutMessage("你已退出本次比赛"))
				return nil
			}
			if num > len(pl.Results) {
				p.sendMessage(msg.NewOutMessage("你还未录入过这一把的成绩"))
				return nil
			}

			res := result.TimeParserS2F(sl[1])
			pkTimerResult.PkResults.Players[idx].Results[num-1] = res
			p.sendMessage(msg.NewOutMessagef("修改第%d把成绩成功, 成绩为: %s", num, result.TimeParserF2S(pkTimerResult.PkResults.Players[idx].Results[num-1])))
			break
		}
	}
	p.Svc.DB.Save(&pkTimerResult)
	return nil
}

// pktimer指令
// 1. 开启pktimer模式， 群聊将不再接受pktimer的之外的指令 pktimer 333 12 pktimer {项目} 把数
// 2. 群员可以加入参与 指令： “加入”
// 3. 开启者可以发送“开始”

func (p *PkTimer) exit(msg types.InMessage) error {
	usr, err := p.getMsgUser(msg)
	if err != nil {
		return err
	}

	pkTimerResult := p.getMessageDBPkTimer(msg)

	hasPlayer := 0
	for _, pl := range pkTimerResult.PkResults.Players {
		if pl.Exit {
			continue
		}
		hasPlayer += 1
	}

	for idx, pl := range pkTimerResult.PkResults.Players {
		if pl.UserId == usr.ID {
			pkTimerResult.PkResults.Players[idx].Exit = true
			pkTimerResult.PkResults.Players[idx].ExitNum = pkTimerResult.PkResults.CurCount
			p.Svc.DB.Save(&pkTimerResult)
			p.sendMessage(msg.NewOutMessagef("%s 退出成功", pl.UserName))

			// 只剩一个的情况下
			if hasPlayer == 1 {
				return p.endPkTimer(msg)
			}
			return nil
		}
	}

	return nil
}

func (p *PkTimer) runMessage(msg types.InMessage) error {
	switch utils2.ReplaceAll(msg.Message, "", " ") {
	case start:
		return p.startPkTimer(msg)
	case end:
		return p.endPkTimer(msg)
	case add:
		return p.in(msg)
	case next:
		return p.next(msg)
	case exit:
		return p.exit(msg)
	default:
		// 判断是否有修改
		if strings.Contains(msg.Message, update) {
			return p.updateResult(msg)
		}
		return p.addResult(msg)
	}
}
