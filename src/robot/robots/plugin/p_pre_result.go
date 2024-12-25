package plugin

import (
	"fmt"
	"slices"
	"strings"
	"time"

	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
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
	msg := RemoveID(message.Message, c.ID())
	msg = c.cutMsg(msg)

	// 0. 判断用户
	usr, err := getUser(c.Svc, message)
	if err != nil || usr.ID == 0 {
		return message.NewOutMessage("用户不存在, 请绑定QQ后重试"), nil
	}
	var group competition.CompetitionGroup
	gms := fmt.Sprintf("%%%v%%", message.GroupID)
	if err = c.Svc.DB.Where("qq_groups LIKE ? or qq_group_uid LIKE ?", gms, gms).First(&group).Error; err != nil {
		return message.NewOutMessage("本群比赛群组未创建或不存在"), nil
	}

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
	msg = strings.TrimLeft(msg, " ")
	return msg
}

func (c *PreResultPlugin) getPreResults(msg string, message types.InMessage, usr user.User) (*types.OutMessage, error) {
	var pre []result.PreResults

	c.Svc.DB.Where("finish = ?", false).Where("user_id = ?", usr.ID).Find(&pre)
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
	var comp competition.Competition
	compDB := c.Svc.DB.Where("group_id = ?", group.ID).Where("is_done = ?", false)
	if msg[0] == '-' {
		first := strings.Index(msg, " ")
		if first == -1 {
			return message.NewOutMessage("格式错误"), nil
		}
		compDB = compDB.Where("id = ?", msg[1:first+1])
		msg = msg[first+1:]
	}
	if err := compDB.Order("created_at DESC").First(&comp).Error; err != nil {
		return message.NewOutMessage("本群未创建比赛或该比赛已经结束"), nil
	}

	// 2. 解析 并编辑录入
	evs := GetEvents(c.Svc, "")
	var pres []result.PreResults

	out := fmt.Sprintf("比赛: %s\n", comp.Name)
	out += fmt.Sprintf("选手: %s\n", usr.Name)

	for _, resStr := range strings.Split(msg, "\n") {
		if len(resStr) == 0 {
			continue
		}
		resStr = utils.ReplaceAll(resStr, "", "(", ")")
		resStr = utils.ReplaceAll(resStr, " ", ",", ";")
		resStr = utils.ReplaceAll(resStr, "[", " [", "  [", "   [")
		resStr = utils.ReplaceAll(resStr, "] ", "]")

		// 获取轮次和项目
		split := strings.Split(resStr, " ")
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
		if err != nil {
			return message.NewOutMessage(fmt.Sprintf("'%s'本次%s比赛未开放`%s` '%v'赛程", resStr, comp.Name, ev.ID, round)), nil
		}

		// todo 晋级机制?
		//if !schedule.FirstRound && !schedule.NoRestrictions && !slices.Contains(schedule.AdvancedToThisRound, usr.ID) {
		//	return message.NewOutMessage(fmt.Sprintf("`%s`,录入错误,你不在晋级名单中", resStr)), nil
		//}

		// 查看是否有旧的数据存在，如果有则覆盖
		var pre result.PreResults
		if err = c.Svc.DB.First(&pre, "comp_id = ? and round_number = ? and user_id = ? and event_id = ?", comp.ID, schedule.RoundNum, usr.ID, ev.ID).Error; err == nil && pre.ID != 0 {
			if pre.Finish { // 已经处理的就重新生成一个
				pre = result.PreResults{}
			}
		}

		// 写入数据

		var results []float64
		for idx, r := range split[1:] {
			if ev.BaseRouteType.RouteMap().Repeatedly && idx == 0 {
				sp2 := strings.Split(r, "/")
				if !strings.Contains(r, "/") || len(sp2) != 2 {
					return message.NewOutMessage("多盲格式不符合'x/y time'的格式"), nil
				}
				results = append(results, utils.GetNum(sp2[0]), utils.GetNum(sp2[1]))
				continue
			}
			results = append(results, result.TimeParserS2F(r))
		}

		preResult := result.PreResults{
			Results: result.Results{
				Model: basemodel.Model{
					//ID: pre.ID,
				},
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
		if err = preResult.Update(); err != nil {
			return message.NewOutMessage(fmt.Sprintf("`%s`成绩存在格式错误: %s", resStr, err.Error())), nil
		}
		pres = append(pres, preResult)

		out += fmt.Sprintf("%s %s (%s / %s)\n", ev.Cn, schedule.Round, preResult.BestString(), preResult.BestAvgString())
	}

	out += "录入成功!\n"
	c.Svc.DB.Save(&pres)
	// todo 破记录播报

	return message.NewOutMessage(out), nil
}
