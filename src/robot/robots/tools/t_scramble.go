package tools

import (
	"fmt"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/robot/types"
)

type TScramble struct {
	Svc *svc.Svc
}

func (t *TScramble) ID() []string {
	var out []string

	var evs []event.Event
	t.Svc.DB.Find(&evs)

	for _, ev := range evs {
		if !ev.IsComp {
			continue
		}
		out = append(out, ev.ID)
	}

	out = append(out, "打乱调试")

	return out
}

func (t *TScramble) Help() string {
	return `打乱`
}

func (t *TScramble) helps() string {
	out := ""
	for i, id := range t.ID() {
		out += fmt.Sprintf("%d. %s\n", i+1, id)
	}
	return out
}

func (t *TScramble) Do(message types.InMessage) (*types.OutMessage, error) {
	m := utils.ReplaceAll(message.Message, "", " ")

	if m == "打乱调试" {
		return message.NewOutMessage(t.Svc.Scramble.Test()), nil
	}

	var ev event.Event
	if err := t.Svc.DB.Where("id = ?", m).First(&ev).Error; err != nil {
		return message.NewOutMessage("打乱不存在\n" + t.helps()), err
	}

	ts := time.Now()
	out := t.Svc.Scramble.Scramble(ev.ID, 1)
	if len(out) == 0 {
		return message.NewOutMessagef("获取打乱错误, 长度0\n"), nil
	}
	use := time.Since(ts)
	msg := ""
	for idx, o := range out {
		msg += fmt.Sprintf("%d. %s\n", idx+1, o)
	}
	msg += fmt.Sprintf("-------------\n耗时：%s\n", use)

	return message.NewOutMessage(msg), nil
}
