package plugin

import (
	"fmt"
	"strings"

	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/robot/types"
)

type PlayerPlugin struct {
	Svc *svc.Svc
}

var _ types.Plugin = &PlayerPlugin{}

func (c *PlayerPlugin) ID() []string {
	return []string{"player", "选手", "玩家"}
}

func (c *PlayerPlugin) Help() string {
	return `获取比赛信息:
1. 选手: 查询自己的最佳成绩
2. 选手-{名称/CubeID}: 可查询其他选手的最佳成绩
* 需要注意是否有绑定QQ号
`
}

func (c *PlayerPlugin) Do(message types.InMessage) (*types.OutMessage, error) {
	msg := RemoveID(message.Message, c.ID())
	msg = utils.ReplaceAll(msg, "", "-", " ")
	var usr user.User
	var err error
	if len(msg) == 0 {
		if message.QQ != 0 {
			err = c.Svc.DB.Where("qq = ?", fmt.Sprintf("%d", message.QQ)).First(&usr).Error
		} else if message.QQBot != "" {
			err = c.Svc.DB.Where("qq_uni_id = ?", message.QQBot).First(&usr).Error
		}
	} else {
		err = c.Svc.DB.Where("name = ?", msg).Or("en_name = ?", msg).Or("cube_id = ?", strings.ToUpper(msg)).First(&usr).Error
	}

	if err != nil {
		return message.NewOutMessage(fmt.Sprintf("查询不到选手 `%s`", msg)), nil
	}
	out := "===== " + usr.Name + " =====\n"
	out += fmt.Sprintf("CubeID: %s\n", usr.CubeID)
	out += fmt.Sprintf("主页: https://cubing.pro/player/%s\n", usr.CubeID)
	out += "\n========================\n"

	best := c.Svc.Cov.SelectBestResultsWithEventSortWithPlayer(usr.CubeID)
	allWca := true
	score := func(ev event.Event) string {
		b, ok := best.Single[ev.ID]
		a, ok2 := best.Avgs[ev.ID]
		if !ok && !ok2 {
			allWca = false
			return ""
		}

		output := ""
		output += fmt.Sprintf("%s (%d) %s", ev.Cn+":", b.Rank, b.BestString())

		if ev.BaseRouteType.RouteMap().Repeatedly {
			return output + "\n"
		}

		if ok2 {
			output += fmt.Sprintf(" | %s (%d)", a.BestAvgString(), a.Rank)
		} else {
			allWca = false
		}
		return output + "\n"
	}

	events := GetEvents(c.Svc, "")
	for _, ev := range events {
		if ev.IsWCA {
			out += score(ev)
		}
	}
	if allWca {
		out += "(该玩家是全项目达成选手 - 大满贯)\n"
	}

	out += "\n========================\n"
	for _, ev := range events {
		if !ev.IsWCA {
			out += score(ev)
		}
	}
	return message.NewOutMessage(out), nil
}
