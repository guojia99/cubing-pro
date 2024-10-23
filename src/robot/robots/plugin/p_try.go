package plugin

import (
	"fmt"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/robot/types"
	"time"
)

//type Plugin interface {
//	ID() []string
//	Help() string
//	Do(message InMessage) (OutMessage, error)
//}

type TryPlugin struct {
	Svc *svc.Svc
}

var _ types.Plugin = &TryPlugin{}

func (t *TryPlugin) ID() []string {
	return []string{"try", "测试"}
}

func (t *TryPlugin) Help() string {
	return "这是一个测试用的指令"
}

func (t *TryPlugin) Do(message types.InMessage) (*types.OutMessage, error) {
	msg := message.Message
	msg = RemoveID(msg, t.ID())
	if len(msg) > 0 {
		return nil, nil
	}
	return message.NewOutMessage(fmt.Sprintf(`==================
	测试: %s
	时间: %s
	发送消息人: %s
	发送群聊: %d
==================
`,
		message.Message,
		time.Now().Local().Format("2006-01-02 15:04:05"),
		message.Name,
		message.GroupID,
	)), nil
}
