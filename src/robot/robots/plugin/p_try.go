package plugin

import (
	"fmt"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/robot/types"
)

type TryPlugin struct {
	Svc *svc.Svc

	checkFn map[string]func(message types.InMessage) (*types.OutMessage, error)
}

var _ types.Plugin = &TryPlugin{}

func (t *TryPlugin) ID() []string {
	t.init()

	var out []string
	for key, _ := range t.checkFn {
		out = append(out, key)
	}
	return out
}

func (t *TryPlugin) Help() string {
	return "这是一个测试用的指令"
}

func (t *TryPlugin) init() {
	t.checkFn = map[string]func(message types.InMessage) (*types.OutMessage, error){
		"try": t._test, "测试": t._test,
		"蛋炒饭": t._egg, "一碗蛋炒饭": t._egg, "一碗炒饭": t._egg, "炒饭": t._egg,
	}

}

func (t *TryPlugin) Do(message types.InMessage) (*types.OutMessage, error) {
	key := message.Message
	fn, ok := t.checkFn[key]
	if !ok {
		return nil, nil
	}
	return fn(message)
}

func (t *TryPlugin) _test(message types.InMessage) (*types.OutMessage, error) {
	return message.NewOutMessage(fmt.Sprintf(`==================
消息: %s
消息长度: %d
时间: %s
发送消息人: %s
发送人QQ: %d
发送人QQBot: %s
发送群聊: %s
==================
`,
		message.Message,
		len(message.Message),
		time.Now().Local().Format("2006-01-02 15:04:05"),
		message.Name,
		message.QQ,
		message.QQBot,
		message.GroupID,
	)), nil
}

func (t *TryPlugin) _egg(message types.InMessage) (*types.OutMessage, error) {
	return message.NewOutMessage("boom!程序爆炸了!"), nil
}
