package plugin

import (
	"fmt"
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
1. 比赛: 可查询当前比赛列表
2. 比赛-{名称/序号}: 可查询比赛的详细信息`
}

func (c *PlayerPlugin) Do(message types.InMessage) (*types.OutMessage, error) {
	msg := RemoveID(message.Message, c.ID())
	msg = utils.ReplaceAll(msg, "", "-", " ")
	var usr user.User
	var err error
	if len(msg) == 0 {
		err = c.Svc.DB.Where("qq = ?", fmt.Sprintf("%d", message.QQ)).First(&usr).Error
	} else {
		err = c.Svc.DB.Where("name = ?", msg).Or("en_name = ?", msg).First(&usr).Error
	}

	if err != nil {
		return message.NewOutMessage("查询不到选手"), nil
	}
	out := "===== " + usr.Name + " =====\n"
	out += fmt.Sprintf("CubeID: %s\n", usr.CubeID)
	out += fmt.Sprintf("主页: https://cubing.pro/player/%s\n", usr.CubeID)
	out += "\n========================\n"

	best := c.Svc.Cov.SelectBestResultsWithEventSortWithPlayer(usr.CubeID)
	score := func(ev event.Event) string {
		b, ok := best.Single[ev.ID]
		a, ok2 := best.Avgs[ev.ID]
		if !ok && !ok2 {
			return ""
		}

		output := ""
		output += fmt.Sprintf("%s (%d) %s", ev.Cn+":", b.Rank, b.BestString())
		if ok2 {
			output += fmt.Sprintf(" | %s (%d)", a.BestAvgString(), a.Rank)
		}
		return output + "\n"
	}
	events := GetEvents(c.Svc, "")
	for _, ev := range events {
		out += score(ev)
	}
	return message.NewOutMessage(out), nil
}
