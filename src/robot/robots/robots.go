package robots

import (
	"context"
	"fmt"
	"github.com/guojia99/cubing-pro/src/robot/robots/plugin"
	"github.com/guojia99/cubing-pro/src/robot/types"
)

func withInMessage(msg types.InMessage, pluginMap map[string]types.Plugin) *types.OutMessage {
	key, p := plugin.CheckPrefix(msg.Message, pluginMap)
	if key == "" {
		return nil
	}
	out, err := p.Do(msg)
	if err != nil {
		fmt.Printf("[%s] 错误 %s\n", key, err)
		return nil
	}
	return out
}

func RunRobot(ctx context.Context, r types.Robot, plugins []types.Plugin) {
	var inCh = make(chan types.InMessage, 128)

	pluginMap := plugin.PluginsMap(plugins)

	// 处理in message
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-inCh:
				out := withInMessage(msg, pluginMap)
				if out == nil {
					continue
				}
				if err := r.SendMessage(*out); err != nil {
					fmt.Printf("发送消息失败 %s\n", err)
				}
			}
		}
	}()

	r.Run(inCh)
}
