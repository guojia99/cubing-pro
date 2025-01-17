package tools

import (
	"fmt"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/svc"
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
	return out
}

func (t *TScramble) Help() string {
	return `打乱`
}

func (t *TScramble) Do(message types.InMessage) (*types.OutMessage, error) {
	ts := time.Now()
	out := t.Svc.Scramble.Scramble(message.Message, 1)
	use := time.Since(ts)

	if len(out) == 0 {
		return message.NewOutMessage("获取打乱失败"), nil
	}
	msg := ""
	for idx, o := range out {
		msg += fmt.Sprintf("%d. %s\n", idx+1, o)
	}
	msg += fmt.Sprintf("-------------\n耗时：%s\n", use)

	return message.NewOutMessage(msg), nil
}
