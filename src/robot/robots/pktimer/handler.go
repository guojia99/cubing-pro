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

const (
	// 超时时间常量
	pkTimeoutDuration = 20 * time.Minute

	// 默认和最大轮次
	defaultCount = 5
	maxCount     = 20

	// 默认精度
	defaultEps = 0.05
)

// 项目精度映射表
var epsMap = map[string]float64{
	"444":   0.018,
	"555":   0.018,
	"minx":  0.02,
	"666":   0.01,
	"777":   0.01,
	"444bf": 0.03,
	"555bf": 0.03,
}

func (p *PkTimer) checkInPkTimer(msg types.InMessage) bool {
	_, err := p.getMsgUser(msg)
	if err != nil {
		return false
	}

	pk := p.getMessageDBPkTimer(msg)
	if pk == nil {
		return false
	}

	// 如果还没开始，允许所有命令
	if !pk.Start {
		return true
	}

	// 检查是否是控制命令
	msgStr := utils2.ReplaceAll(msg.Message, "", " ")
	switch msgStr {
	case start, exit, end, add, next:
		return true
	}

	// 已开始，检查用户是否在参与PK中
	return p.isPlayerInPk(pk, msg.QQBot)
}

// isPlayerInPk 检查玩家是否在PK中
func (p *PkTimer) isPlayerInPk(pk *pktimerDB.PkTimerResult, qqBot string) bool {
	for _, pl := range pk.PkResults.Players {
		if pl.QQBot == qqBot && !pl.Exit {
			return true
		}
	}
	return false
}

func (p *PkTimer) initPkTimer(msg types.InMessage) error {
	// 检查并清理超时的PK
	if err := p.cleanupExpiredPk(msg); err != nil {
		return err
	}

	// 处理重新加载上一轮PK
	if strings.Contains(msg.Message, reload) {
		return p.reloadLastPk(msg)
	}

	return p.createNewPkTimer(msg)
}

// cleanupExpiredPk 清理过期的PK
func (p *PkTimer) cleanupExpiredPk(msg types.InMessage) error {
	var pkTimerResult pktimerDB.PkTimerResult
	if err := p.Svc.DB.Where("group_id = ? AND running = ?", msg.GroupIDStr(), true).
		First(&pkTimerResult).Error; err != nil {
		return nil // 没有运行中的PK，直接返回
	}

	if time.Since(pkTimerResult.LastRunning) < pkTimeoutDuration {
		p.sendMessage(msg.NewOutMessage("请结束上一次pk后再开始pk"))
		return nil
	}

	// 标记为已结束
	pkTimerResult.Running = false
	return p.Svc.DB.Save(&pkTimerResult).Error
}

// reloadLastPk 重新加载上一轮PK
func (p *PkTimer) reloadLastPk(msg types.InMessage) error {
	var lastRunning pktimerDB.PkTimerResult
	if err := p.Svc.DB.Order("created_at DESC").
		Where("group_id = ?", msg.GroupIDStr()).
		First(&lastRunning).Error; err != nil {
		p.sendMessage(msg.NewOutMessage("上一次pk不存在, 请创建后再进行"))
		return nil
	}

	newPK := pktimerDB.PkTimerResult{
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

	// 重置所有玩家的成绩
	for idx := range newPK.PkResults.Players {
		newPK.PkResults.Players[idx].Results = make([]float64, 0)
		newPK.PkResults.Players[idx].Average = result.DNF
		newPK.PkResults.Players[idx].Best = result.DNF
	}

	if err := p.Svc.DB.Save(&newPK).Error; err != nil {
		return err
	}

	p.sendMessage(msg.NewOutMessage(getIniterMessage(&newPK)))
	return nil
}

func (p *PkTimer) createNewPkTimer(msg types.InMessage) error {
	usr, err := p.getMsgUser(msg)
	if err != nil {
		p.sendMessage(msg.NewOutMessagef("用户`%d | %s`未绑定，请绑定Cubing Pro后再使用", msg.QQ, msg.QQBot))
		return err
	}

	// 解析命令参数
	ev, count, eps, err := p.parsePkTimerCommand(msg.Message)
	if err != nil {
		p.sendMessage(msg.NewOutMessage("请输入: `pktimer 333 5 1`来代表pk5次3速,精度1%"))
		return nil
	}

	// 查找项目
	eve, err := p.findEvent(ev)
	if err != nil || eve.ID == "" {
		p.sendMessage(msg.NewOutMessage("不支持的项目"))
		return nil
	}

	// 检查项目是否支持
	rm := eve.BaseRouteType.RouteMap()
	if rm.Repeatedly || !eve.IsWCA {
		p.sendMessage(msg.NewOutMessage("不支持的项目"))
		return nil
	}

	// 创建新的PK
	newPKTimerResult := pktimerDB.PkTimerResult{
		GroupID:     msg.GroupIDStr(),
		Running:     true,
		LastRunning: time.Now(),
		StartPerson: usr.Name,
		PkResults: pktimerDB.PkResults{
			Players: []pktimerDB.Player{
				{
					QQ:       msg.QQ,
					QQBot:    msg.QQBot,
					UserName: usr.Name,
					UserId:   usr.ID,
					Results:  make([]float64, 0),
				},
			},
			Event:        eve,
			Count:        count,
			CurCount:     0,
			FirstMessage: msg,
		},
		Eps: eps,
	}

	if err := p.Svc.DB.Save(&newPKTimerResult).Error; err != nil {
		return err
	}

	p.sendMessage(msg.NewOutMessage(getIniterMessage(&newPKTimerResult)))
	return nil
}

// parsePkTimerCommand 解析PK命令参数
func (p *PkTimer) parsePkTimerCommand(message string) (event string, count int, eps float64, err error) {
	m := utils2.ReplaceAll(message, "", key)
	slice := utils2.Split(m, " ")

	if len(slice) == 0 || len(slice) >= 4 {
		return "", 0, 0, fmt.Errorf("invalid command format")
	}

	event = slice[0]

	// 解析轮次
	if len(slice) >= 2 {
		count, _ = strconv.Atoi(slice[1])
	}
	if count <= 0 {
		count = defaultCount
	}
	if count > maxCount {
		count = maxCount
	}

	// 解析精度
	eps = defaultEps
	if o, ok := epsMap[event]; ok {
		eps = o
	}
	if len(slice) >= 3 {
		if epsInt, err := strconv.Atoi(slice[2]); err == nil {
			eps = float64(epsInt) / 100.0
		}
	}

	return event, count, eps, nil
}

// findEvent 查找项目
func (p *PkTimer) findEvent(ev string) (event.Event, error) {
	var events []event.Event
	if err := p.Svc.DB.Where("is_wca = ?", true).Find(&events).Error; err != nil {
		return event.Event{}, err
	}

	for _, e := range events {
		// 精确匹配
		if e.ID == ev || e.Name == ev || e.Cn == ev {
			return e, nil
		}
		// 检查其他名称
		if e.OtherNames != "" {
			otherNames := utils2.Split(e.OtherNames, ";")
			for _, name := range otherNames {
				if name == ev {
					return e, nil
				}
			}
		}
	}

	return event.Event{}, fmt.Errorf("event not found")
}

func (p *PkTimer) in(msg types.InMessage) error {
	pkTimerResult := p.getMessageDBPkTimer(msg)
	if pkTimerResult == nil {
		return fmt.Errorf("pk timer not found")
	}

	if pkTimerResult.Start {
		p.sendMessage(msg.NewOutMessage("本次pk已经开始，请下一次pk及时加入"))
		return nil
	}

	usr, err := p.getMsgUser(msg)
	if err != nil {
		p.sendMessage(msg.NewOutMessagef("用户`%d | %s`未绑定，请绑定Cubing Pro后再使用", msg.QQ, msg.QQBot))
		return err
	}

	// 检查是否已加入
	if p.isUserInPlayers(pkTimerResult.PkResults.Players, usr.ID) {
		p.sendMessage(msg.NewOutMessage("你已经加入本次pk"))
		p.sendMessage(msg.NewOutMessage(getIniterMessage(pkTimerResult)))
		return nil
	}

	// 添加新玩家
	player := pktimerDB.Player{
		Index:    len(pkTimerResult.PkResults.Players) + 1,
		QQ:       msg.QQ,
		QQBot:    msg.QQBot,
		UserName: usr.Name,
		UserId:   usr.ID,
		Results:  make([]float64, 0),
	}
	pkTimerResult.PkResults.Players = append(pkTimerResult.PkResults.Players, player)
	pkTimerResult.LastRunning = time.Now()
	p.Svc.DB.Save(&pkTimerResult)
	p.sendMessage(msg.NewOutMessage(getIniterMessage(pkTimerResult)))
	return nil
}

// isUserInPlayers 检查用户是否已在玩家列表中
func (p *PkTimer) isUserInPlayers(players []pktimerDB.Player, userID uint) bool {
	for _, pl := range players {
		if pl.UserId == userID {
			return true
		}
	}
	return false
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
	if pkTimerResult == nil {
		return fmt.Errorf("pk timer not found")
	}

	// 计算所有玩家的最佳和平均成绩
	rm := pkTimerResult.PkResults.Event.BaseRouteType.RouteMap()
	for idx := range pkTimerResult.PkResults.Players {
		pl := &pkTimerResult.PkResults.Players[idx]
		best, avg := result.GetBestAndAvg(pl.Results, rm)
		pl.Best = best
		pl.Average = avg
	}

	// 保存最终成绩
	if err := p.Svc.DB.Save(&pkTimerResult).Error; err != nil {
		return err
	}

	// 发送结果消息
	p.sendMessage(msg.NewOutMessage(p.endPKTimerMessage(pkTimerResult)))
	_ = p.sendPackerMessage(msg, false)

	// 标记为已结束
	pkTimerResult.Running = false
	p.Svc.DB.Save(&pkTimerResult)
	return nil
}

func (p *PkTimer) sendScrambleMessage(msg types.InMessage) {
	pkTimerResult := p.getMessageDBPkTimer(msg)
	if pkTimerResult == nil {
		return
	}

	pkTimerResult.PkResults.CurCount++
	sc, err := p.Svc.Scramble.ScrambleWithEvent(pkTimerResult.PkResults.Event, 1)
	if err != nil || len(sc) == 0 {
		log.Error(err)
		return
	}

	pkTimerResult.LastRunning = time.Now()
	p.Svc.DB.Save(&pkTimerResult)

	scrMsg := fmt.Sprintf("本次pk第%d / %d把:\n %s",
		pkTimerResult.PkResults.CurCount, pkTimerResult.PkResults.Count, sc[0])

	img, err := p.Svc.Scramble.Image(sc[0], pkTimerResult.PkResults.Event.ID)
	if err != nil {
		p.sendMessage(msg.NewOutMessage(scrMsg))
		return
	}

	filePath := path.Join(os.TempDir(), fmt.Sprintf("pktimer_%d_%d.jpg",
		time.Now().UnixNano(), pkTimerResult.PkResults.CurCount))
	_ = os.WriteFile(filePath, []byte(img), 0644)
	p.sendMessage(msg.NewOutMessageWithImage(scrMsg, filePath))
}

func (p *PkTimer) next(msg types.InMessage) error {
	pkTimerResult := p.getMessageDBPkTimer(msg)
	if pkTimerResult == nil {
		return fmt.Errorf("pk timer not found")
	}

	// 处理没有成绩的玩家，记录为DNS
	curCount := pkTimerResult.PkResults.CurCount
	hasDns := false
	for idx := range pkTimerResult.PkResults.Players {
		pl := &pkTimerResult.PkResults.Players[idx]
		if !pl.Exit && len(pl.Results) != curCount {
			pl.Results = append(pl.Results, result.DNF)
			if !hasDns {
				hasDns = true
			}
		}
	}

	if hasDns {
		p.sendMessage(msg.NewOutMessage("将进行下一把, 未录入成绩的玩家本把成绩记录为DNS"))
	}

	// 检查是否所有轮次已完成
	if pkTimerResult.PkResults.Count == pkTimerResult.PkResults.CurCount {
		return p.endPkTimer(msg)
	}

	// 保存状态
	if err := p.Svc.DB.Save(&pkTimerResult).Error; err != nil {
		return err
	}

	_ = p.sendPackerMessage(msg, true)
	p.sendScrambleMessage(msg)
	return nil
}

func (p *PkTimer) addResult(msg types.InMessage) error {
	pkTimerResult := p.getMessageDBPkTimer(msg)
	if pkTimerResult == nil || !pkTimerResult.Start {
		return nil
	}

	// 检查用户是否在PK中
	usr, err := p.getMsgUser(msg)
	if err != nil {
		return err
	}

	playerIdx := p.findPlayerIndex(pkTimerResult.PkResults.Players, usr.ID, msg.QQBot, msg.QQ)
	if playerIdx == -1 {
		return nil // 用户不在PK中
	}

	// 解析成绩
	res := result.TimeParserS2F(msg.Message)
	if res == result.UNT {
		return nil
	}

	// 更新或添加成绩
	player := &pkTimerResult.PkResults.Players[playerIdx]
	roundNum := pkTimerResult.PkResults.CurCount

	if len(player.Results) == roundNum {
		// 更新已有成绩
		player.Results[roundNum-1] = res
		p.sendMessage(msg.NewOutMessagef("你的第 %d 把 PK 成绩更新成功: %s | (%d/%d)",
			roundNum, result.TimeParserF2S(res), roundNum, pkTimerResult.PkResults.Count))
	} else {
		// 添加新成绩
		player.Results = append(player.Results, res)
		p.sendMessage(msg.NewOutMessagef("你的第 %d 把 PK 成绩记录成功: %s | (%d/%d)",
			len(player.Results), result.TimeParserF2S(res), len(player.Results), pkTimerResult.PkResults.Count))
	}

	if err := p.Svc.DB.Save(&pkTimerResult).Error; err != nil {
		return err
	}

	// 检查是否所有玩家都已完成本轮
	if p.allPlayersCompleted(pkTimerResult) {
		return p.next(msg)
	}

	return nil
}

// findPlayerIndex 查找玩家索引
func (p *PkTimer) findPlayerIndex(players []pktimerDB.Player, userID uint, qqBot string, qq int64) int {
	for idx, pl := range players {
		if pl.UserId == userID || pl.QQBot == qqBot || pl.QQ == qq {
			return idx
		}
	}
	return -1
}

// allPlayersCompleted 检查是否所有玩家都已完成当前轮次
func (p *PkTimer) allPlayersCompleted(pkTimerResult *pktimerDB.PkTimerResult) bool {
	curCount := pkTimerResult.PkResults.CurCount
	for _, pl := range pkTimerResult.PkResults.Players {
		if pl.Exit {
			continue
		}
		if len(pl.Results) != curCount {
			return false
		}
	}
	return true
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

	roundNum, err := strconv.Atoi(sl[0])
	if err != nil || roundNum <= 0 {
		p.sendMessage(msg.NewOutMessage("修改成绩格式为: `修改 1 1.23`, 代表第一把修改成绩为1.23"))
		return nil
	}

	usr, err := p.getMsgUser(msg)
	if err != nil {
		return err
	}

	pkTimerResult := p.getMessageDBPkTimer(msg)
	if pkTimerResult == nil {
		return fmt.Errorf("pk timer not found")
	}

	if roundNum > pkTimerResult.PkResults.CurCount {
		p.sendMessage(msg.NewOutMessage("该轮次还未开始，无法修改"))
		return nil
	}

	playerIdx := p.findPlayerIndexByUserID(pkTimerResult.PkResults.Players, usr.ID)
	if playerIdx == -1 {
		return nil
	}

	player := &pkTimerResult.PkResults.Players[playerIdx]
	if player.Exit {
		p.sendMessage(msg.NewOutMessage("你已退出本次比赛"))
		return nil
	}

	if roundNum > len(player.Results) {
		p.sendMessage(msg.NewOutMessage("你还未录入过这一把的成绩"))
		return nil
	}

	// 解析并更新成绩
	res := result.TimeParserS2F(sl[1])
	if res == result.UNT {
		p.sendMessage(msg.NewOutMessage("成绩格式错误"))
		return nil
	}

	player.Results[roundNum-1] = res

	if err := p.Svc.DB.Save(&pkTimerResult).Error; err != nil {
		return err
	}

	p.sendMessage(msg.NewOutMessagef("修改第%d把成绩成功, 成绩为: %s",
		roundNum, result.TimeParserF2S(player.Results[roundNum-1])))
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
	if pkTimerResult == nil {
		return fmt.Errorf("pk timer not found")
	}

	// 统计未退出的玩家数量
	activePlayerCount := p.countActivePlayers(pkTimerResult.PkResults.Players)

	// 标记玩家退出
	playerIdx := p.findPlayerIndexByUserID(pkTimerResult.PkResults.Players, usr.ID)
	if playerIdx == -1 {
		return nil
	}

	player := &pkTimerResult.PkResults.Players[playerIdx]
	player.Exit = true
	player.ExitNum = pkTimerResult.PkResults.CurCount

	if err := p.Svc.DB.Save(&pkTimerResult).Error; err != nil {
		return err
	}

	p.sendMessage(msg.NewOutMessagef("%s 退出成功", player.UserName))

	// 如果只剩一个玩家，自动结束PK
	if activePlayerCount <= 1 {
		return p.endPkTimer(msg)
	}

	return nil
}

// countActivePlayers 统计活跃玩家数量
func (p *PkTimer) countActivePlayers(players []pktimerDB.Player) int {
	count := 0
	for _, pl := range players {
		if !pl.Exit {
			count++
		}
	}
	return count
}

//func (p *PkTimer) out(msg types.InMessage) error {
//	pkTimerResult := p.getMessageDBPkTimer(msg)
//	if pkTimerResult == nil {
//		return fmt.Errorf("pk timer not found")
//	}
//
//}

// findPlayerIndexByUserID 根据用户ID查找玩家索引
func (p *PkTimer) findPlayerIndexByUserID(players []pktimerDB.Player, userID uint) int {
	for idx, pl := range players {
		if pl.UserId == userID {
			return idx
		}
	}
	return -1
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
	//case out:
	//	return p.out(msg)
	case update:
		return p.updateResult(msg)
	default:
		return p.addResult(msg)
	}
}
