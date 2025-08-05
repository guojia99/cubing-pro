package tools

import (
	"fmt"
	"os"
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
		out = append(out, ev.ID, ev.Name)
		data := utils.Split(ev.OtherNames, ";")
		out = append(out, data...)
	}

	out = append(out, "打乱调试")
	return utils.RemoveDuplicates(out)
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

	var evs []event.Event
	var curEv event.Event

	t.Svc.DB.Where("is_comp = ?", true).Find(&evs)

	// 检查
	for _, e := range evs {
		if e.ID == m || e.Name == m {
			curEv = e
			break
		}

		sl := utils.Split(e.OtherNames, ";")
		for _, v := range sl {
			if v == m {
				curEv = e
			}
		}
	}

	if curEv.ID == "" {
		return message.NewOutMessage("打乱指令不存在"), nil
	}

	out, err := t.Svc.Scramble.ScrambleWithEvent(curEv, 1)
	if err != nil {
		return message.NewOutMessage(err.Error()), nil
	}
	if len(out) == 0 {
		return message.NewOutMessagef("获取打乱错误, 长度0\n"), nil
	}

	if len(out) == 1 {
		img, err := t.Svc.Scramble.Image(out[0], curEv.ID)
		if err != nil {
			return message.NewOutMessage(out[0]), nil
		}

		filePath := fmt.Sprintf("/%s/_scramble_%d.jpg", os.TempDir(), time.Now().UnixNano())
		_ = os.WriteFile(filePath, []byte(img), 0644)
		return message.NewOutMessageWithImage(filePath), nil
	}

	msg := ""
	for idx, o := range out {
		msg += fmt.Sprintf("%d. %s\n", idx+1, o)
	}
	return message.NewOutMessage(msg), nil
}
