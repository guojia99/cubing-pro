package plugin

import (
	"fmt"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/robot/types"
	"strings"
)

type CompsPlugin struct {
	Svc *svc.Svc
}

var _ types.Plugin = &CompsPlugin{}

func (c *CompsPlugin) ID() []string {
	return []string{"comp", "comps", "比赛", "比赛列表"}
}

func (c *CompsPlugin) Help() string {
	return `获取比赛信息:
1. 比赛: 可查询当前比赛列表
2. 比赛-{名称/序号}: 可查询比赛的详细信息`
}

func (c *CompsPlugin) Do(message types.InMessage) (*types.OutMessage, error) {
	if strings.Contains(message.Message, "comps") || strings.Contains(message.Message, "比赛列表") {
		return c.comps(message)
	}
	return c.comp(message)
}

func (c *CompsPlugin) comp(message types.InMessage) (*types.OutMessage, error) {
	msg := RemoveID(message.Message, c.ID())
	msg = utils.ReplaceAll(msg, "", "-", " ")

	var id = 0
	if number := utils.GetNumbers(msg); len(number) > 0 {
		id = int(number[0])
	}
	var comp competition.Competition
	var err error
	if id == 0 {
		err = c.Svc.DB.Where("status = ?", competition.Running).Order("created_at DESC").First(&comp).Error
	} else {
		err = c.Svc.DB.Where("status = ?", competition.Running).Where("id = ?", id).First(&comp).Error
	}
	if err != nil {
		return message.NewOutMessage(fmt.Sprintf("找不到比赛%d", id)), nil
	}

	var out = fmt.Sprintf("%s\n\n", comp.Name)

	out += fmt.Sprintf("状态: %s\n", comp.StatusName())
	out += fmt.Sprintf("比赛时间: %s ~ %s\n", comp.CompStartTime.Format("20060102"), comp.CompEndTime.Format("20060102"))

	out += "比赛项目: "
	var events = GetEvents(c.Svc, comp.EventMin)
	for _, ev := range events {
		out += ev.Cn + " "
	}

	out += "\n"

	return message.NewOutMessage(out), nil
}

func (c *CompsPlugin) comps(message types.InMessage) (*types.OutMessage, error) {
	var comps []competition.Competition

	var page = 0
	numbers := utils.GetNumbers(message.Message)
	if len(numbers) > 0 {
		page = int(numbers[0])
	}
	offset := (page - 1) * 10

	if err := c.Svc.DB.Where("status = ?", competition.Running).Order("created_at DESC").Offset(offset).Limit(10).Find(&comps).Error; err != nil {
		return nil, err
	}
	out := "===== 比赛列表 =====\n"
	for _, comp := range comps {
		out += fmt.Sprintf("%d. [%s] [%s] %s\n", comp.ID, comp.StatusName(), comp.CompStartTime.Format("20060102"), comp.Name)
	}
	return message.NewOutMessage(out), nil
}
