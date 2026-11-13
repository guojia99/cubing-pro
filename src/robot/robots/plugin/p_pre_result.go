package plugin

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/groupcomp"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/robot/types"
)

type PreResultPlugin struct {
	Svc *svc.Svc
}

func (c *PreResultPlugin) ID() []string {
	return []string{"pre_result", "录入", "预录入", "添加成绩"}
}

func (c *PreResultPlugin) Help() string {
	return `录入群赛成绩:
1. "录入 333 1 2 3 4 5"
2. 允许多个成绩同时录入: "录入 333 1 2 3 4 5 / 444 1 2 3 4 5"
3. 允许DNF, DNS: "录入 333 DNS DNF d s D"
4. 录入某个轮次的: "录入 333[1] 1 2 3 4 5" 注意, 333 后面不能有空格
5. 录入某场比赛: "录入-{比赛ID} 333 1 2 3 4 5"

* 注意, 请不要输入中文字符
`
}

func (c *PreResultPlugin) Do(message types.InMessage) (*types.OutMessage, error) {
	msg := strings.TrimSpace(message.Message)
	for _, id := range c.ID() {
		if strings.HasPrefix(msg, id) {
			msg = strings.TrimPrefix(msg, id)
			break
		}
	}
	msg = c.cutMsg(msg)

	// 0. 判断用户
	usr, err := getUser(c.Svc, message)
	if err != nil || usr.ID == 0 {
		return message.NewOutMessage("用户不存在, 请绑定QQ后重试"), nil
	}
	var groups []competition.CompetitionGroup
	if err = c.Svc.DB.Order("id ASC").Find(&groups).Error; err != nil {
		return message.NewOutMessage("查询比赛群组失败"), nil
	}
	var matched []competition.CompetitionGroup
	for _, g := range groups {
		field := g.QQGroups
		if message.QQ == 0 && message.QQBot != "" {
			field = g.QQGroupUid
		}
		if groupcomp.MatchesGroup(field, fmt.Sprint(message.GroupID)) {
			matched = append(matched, g)
		}
	}
	if len(matched) != 1 {
		return message.NewOutMessage("本群比赛群组未创建或不存在"), nil
	}
	group := matched[0]

	if len(msg) == 0 {
		return c.getPreResults(msg, message, usr)
	}
	return c.setResults(msg, message, usr, group)
}

func (c *PreResultPlugin) cutMsg(msg string) string {
	msg = utils.ReplaceAll(msg, ":", "：")
	msg = utils.ReplaceAll(msg, ",", "，", ", ")
	msg = utils.ReplaceAll(msg, ".", "。")
	msg = utils.ReplaceAll(msg, "[", "【", "〔", "〈", "［", "{")
	msg = utils.ReplaceAll(msg, "]", "】", "〕", "〉", "］", "}")
	msg = utils.ReplaceAll(msg, "(", "（")
	msg = utils.ReplaceAll(msg, ")", "）")

	// 最后一步
	msg = strings.TrimSpace(msg)
	return msg
}

func (c *PreResultPlugin) getPreResults(msg string, message types.InMessage, usr user.User) (*types.OutMessage, error) {
	var pre []result.PreResults

	if err := c.Svc.DB.Where("finish = ?", false).Where("user_id = ?", usr.ID).Find(&pre).Error; err != nil {
		return message.NewOutMessage("查询待审核成绩失败"), nil
	}
	if len(pre) == 0 {
		return message.NewOutMessage("当前没有需要审核的成绩"), nil
	}
	out := fmt.Sprintf("选手: %s\n", usr.Name)
	out += "待审核成绩列表如下:\n"

	maxD := time.Duration(0)
	for _, val := range pre {
		out += fmt.Sprintf("%d.[%d] %s %s (%s / %s)\n",
			val.ID, val.CompetitionID, val.EventName, val.RoundName, val.BestString(), val.BestAvgString())
		if d := time.Since(val.CreatedAt); d > maxD {
			maxD = d
		}
	}

	out += fmt.Sprintf("最长审核已等待 %s\n", utils.DurationToChinese(maxD))

	return message.NewOutMessage(out), nil
}

func (c *PreResultPlugin) setResults(msg string, message types.InMessage, usr user.User, group competition.CompetitionGroup) (*types.OutMessage, error) {
	// 成绩结构:
	// {-CompID} ...{Event}{[RoundNum]} {Results}

	// 1. 判断是否有比赛
	var explicitID uint
	if msg[0] == '-' {
		first := strings.Index(msg, " ")
		if first == -1 {
			return message.NewOutMessage("格式错误"), nil
		}
		id, err := strconv.ParseUint(msg[1:first], 10, 32)
		if err != nil || id == 0 {
			return message.NewOutMessage("比赛ID格式错误"), nil
		}
		explicitID = uint(id)
		msg = msg[first+1:]
	}
	comp, err := groupcomp.SelectRobotCompetition(c.Svc.DB, group.ID, explicitID, time.Now())
	if err != nil {
		if explicitID == 0 {
			return message.NewOutMessage("本群没有可默认录入的已审批未结束群赛；请使用 录入-ID 指定已审批的历史比赛"), nil
		}
		return message.NewOutMessage("指定比赛不属于本群、未审批或不允许预录入，不能回退其他比赛"), nil
	}

	// 2. 解析 并编辑录入
	evs := GetEvents(c.Svc, "")
	var pres []result.PreResults

	out := fmt.Sprintf("比赛: %s\n", comp.Name)
	out += fmt.Sprintf("选手: %s\n", usr.Name)

	// 3. 最佳成绩获取
	playerBestResult, _ := c.Svc.Cov.PlayerBestResult(usr.ID, []string{}, nil) // 玩家最佳成绩
	allBest, _ := c.Svc.Cov.SelectAllPlayerBestResult()

	batch := utils.Split(msg, "\n")
	if comp.IsGroupCompetition() {
		batch = groupcomp.SplitBatch(msg)
	}
	for _, resStr := range batch {
		if len(resStr) == 0 {
			continue
		}
		if comp.IsGroupCompetition() {
			resStr = strings.NewReplacer("(", "[", ")", "]").Replace(resStr)
		} else {
			resStr = utils.ReplaceAll(resStr, "", "(", ")")
		}
		resStr = utils.ReplaceAll(resStr, " ", ",", ";")
		resStr = utils.ReplaceAll(resStr, "[", " [", "  [", "   [")
		resStr = utils.ReplaceAll(resStr, "] ", "]")

		// 获取轮次和项目
		split := utils.Split(resStr, " ")
		if comp.IsGroupCompetition() {
			split = strings.Fields(resStr)
		}
		split = slices.DeleteFunc(split, func(s string) bool { return len(s) == 0 })
		if len(split) <= 1 {
			return message.NewOutMessage(fmt.Sprintf("存在成绩录入格式错误, 或有空数据`%s`", resStr)), nil
		}
		ev, round, _, err := GetMessageEvent(evs, split[0])
		if err != nil {
			return message.NewOutMessage(fmt.Sprintf("`%s` 存在错误:%s %s", resStr, err.Error(), split[0])), nil
		}
		// 确认赛程
		evWithComp, ok := comp.EventMap()[ev.ID]
		if !ok {
			return message.NewOutMessage(fmt.Sprintf("'%s'本次%s比赛未开放`%s`赛程", resStr, comp.Name, ev.ID)), nil
		}
		schedule, err := evWithComp.CurRunningSchedule(round, nil)
		if comp.IsGroupCompetition() {
			_, schedule, err = groupcomp.Validate(comp, ev.ID, round, usr.ID, time.Now(), true, true)
		}
		if err != nil {
			return message.NewOutMessage(fmt.Sprintf("'%s'本次%s比赛未开放`%s` '%v'赛程", resStr, comp.Name, ev.ID, round)), nil
		}

		// 写入数据
		var results []float64
		if comp.IsGroupCompetition() {
			results, err = groupcomp.ParseValues(split[1:], evWithComp.EventRoute)
		} else {
			for idx, r := range split[1:] {
				if ev.BaseRouteType.RouteMap().Repeatedly && idx == 0 {
					parts := utils.Split(r, "/")
					if !strings.Contains(r, "/") || len(parts) != 2 {
						return message.NewOutMessage("多盲格式不符合'x/y time'的格式"), nil
					}
					results = append(results, utils.GetNum(parts[0]), utils.GetNum(parts[1]))
					continue
				}
				results = append(results, result.TimeParserS2F(r))
			}
		}
		if err != nil {
			return message.NewOutMessage(err.Error()), nil
		}

		preResult := result.PreResults{
			Results: result.Results{
				CompetitionID:   comp.ID,
				CompetitionName: comp.Name,
				Round:           schedule.Round,
				RoundNumber:     schedule.RoundNum,
				PersonName:      usr.Name,
				CubeID:          usr.CubeID,
				UserID:          usr.ID,
				Result:          results,
				//Penalty:         req.Penalty,
				EventID:    evWithComp.EventID,
				EventName:  ev.Cn,
				EventRoute: evWithComp.EventRoute,
			},
			CompsName: comp.Name,
			RoundName: schedule.Round,
			Recorder:  usr.Name,
			Source:    fmt.Sprintf("robot-group-%d", group.ID),
		}
		if comp.IsGroupCompetition() {
			preResult.Result = result.UpdateOrgResult(preResult.Result, evWithComp.EventRoute, schedule.Cutoff, schedule.CutoffNumber, schedule.TimeLimit)
		}
		if err = preResult.Update(); err != nil {
			return message.NewOutMessage(fmt.Sprintf("`%s`成绩存在格式错误: %s", resStr, err.Error())), nil
		}
		pres = append(pres, preResult)

		out += fmt.Sprintf("%s %s (%s / %s)\n", ev.Cn, schedule.Round, preResult.BestString(), preResult.BestAvgString())

		// todo 破记录播报
		// 全部记录
		allBestSingle, allBestSingleOk := allBest.Single[preResult.EventID]
		allBestAvg, allBestAvgOk := allBest.Avgs[preResult.EventID]
		hasBest, hasAvg := preResult.DBest(), preResult.DAvg()
		isSingleRound := preResult.EventRoute.RouteMap().Repeatedly || preResult.EventRoute.RouteMap().Rounds == 1

		switch {
		case !allBestSingleOk && !allBestAvgOk && !hasBest && !hasAvg:
			out += fmt.Sprintf("(该成绩是第一个成功有单次和平均有效成绩)\n")
			continue
		case !allBestSingleOk && !hasBest:
			out += "(该成绩是第一个成功有单次成绩)\n"
		case !allBestAvgOk && !hasAvg && !isSingleRound:
			out += "(该成绩是第一个成功有平均成绩)\n"
		}

		// 判断是否打破记录
		isBest := allBestSingleOk && preResult.IsBest(allBestSingle)
		isAvgBest := allBestAvgOk && preResult.IsBestAvg(allBestAvg)

		switch {
		case isBest && isAvgBest && !isSingleRound:
			out += fmt.Sprintf("(该成绩双刷了`%s | %s` 的历史最佳成绩 (%s / %s))\n", allBestSingle.PersonName, allBestAvg.PersonName, allBestSingle.BestString(), allBestAvg.BestAvgString())
			continue
		case isBest:
			out += fmt.Sprintf("(该成绩打破了`%s` 的历史单次最佳成绩 %s)\n", allBestSingle.PersonName, allBestSingle.BestString())
		case isAvgBest && !isSingleRound:
			out += fmt.Sprintf("(该成绩打破了`%s` 的历史平均最佳成绩 %s)\n", allBestAvg.PersonName, allBestAvg.BestAvgString())
		}

		// 个人记录
		pbBestSingle, pbBestSingleOk := playerBestResult.Single[preResult.EventID]
		pbBestAvg, pbBestAvgOk := playerBestResult.Avgs[preResult.EventID]
		hasBestSinglePb := pbBestSingleOk && preResult.IsBest(pbBestSingle)
		hasBestAvgPb := pbBestAvgOk && preResult.IsBestAvg(pbBestAvg)
		switch {
		case hasBestSinglePb && hasBestAvgPb && !isSingleRound:
			out += fmt.Sprintf("(该成绩双刷了自己的历史最佳成绩 (%s / %s))\n", pbBestSingle.BestString(), pbBestAvg.BestAvgString())
			continue
		case hasBestSinglePb:
			out += fmt.Sprintf("(该成绩打破了自己的历史单次最佳成绩 %s)\n", pbBestSingle.BestString())
		case hasBestAvgPb && !isSingleRound:
			out += fmt.Sprintf("(该成绩打破了自己的历史平均最佳成绩 %s)\n", pbBestAvg.BestAvgString())
		}
	}

	out += "录入成功!\n"
	var saveErr error
	if comp.IsGroupCompetition() {
		saveErr = groupcomp.SavePending(c.Svc.DB, pres)
	} else {
		saveErr = c.Svc.DB.Save(&pres).Error
	}
	if err := saveErr; err != nil {
		return message.NewOutMessage("录入失败: " + err.Error()), nil
	}

	return message.NewOutMessage(out), nil
}
