package robots

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"time"

	"github.com/donnie4w/go-logger/logger"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/robot/robots/pktimer"
	"github.com/guojia99/cubing-pro/src/robot/robots/plugin"
	"github.com/guojia99/cubing-pro/src/robot/types"
)

type withFn = func(msg types.InMessage, pluginMap map[string]types.Plugin) (*types.OutMessage, error)

func withInMessage(msg types.InMessage, pluginMap map[string]types.Plugin) (*types.OutMessage, error) {
	isHelp := strings.Contains(msg.Message, "帮助") || strings.Contains(msg.Message, "help")

	msg.Message = utils.ReplaceAll(msg.Message, "", "帮助", "help")

	key, p := plugin.CheckPrefix(msg.Message, pluginMap)
	if key != "" && isHelp {
		return msg.NewOutMessage(p.Help()), nil
	}
	if key == "" {
		return nil, nil
	}
	return p.Do(msg)
}

func withRandCopy(msg types.InMessage, pluginMap map[string]types.Plugin) (*types.OutMessage, error) {
	ra := rand.New(rand.NewSource(time.Now().UnixNano()))
	pr, lens := 0.01, 35
	if ra.Float64() < pr && len(msg.Message) < lens {
		return msg.NewOutMessage(msg.Message), nil
	}
	return nil, nil
}

func RunRobot(ctx context.Context, r types.Robot, plugins []types.Plugin, pkTimerClient *pktimer.PkTimer) {
	var inCh = make(chan types.InMessage, 128)

	pluginMap := plugin.PluginsMap(plugins)

	var str = ""
	for key, _ := range pluginMap {
		str += fmt.Sprintf("%s|", key)
	}
	logger.Infof("[Robot] key => `%s`", str)

	withFns := []withFn{
		withInMessage,
		withRandCopy,
	}

	// 处理in message
	go func() {
		for {
			f := func() {
				defer func() {
					if err := recover(); err != nil {
						logger.Error("[Robot] 错误", err)
						buf := make([]byte, 1024*24)
						n := runtime.Stack(buf, false)
						logger.Errorf("[Robot] 堆栈错误\n %s \n", buf[:n])
					}
				}()

				select {
				case <-ctx.Done():
					return
				case msg := <-inCh:
					inPkTimer := pkTimerClient.WithInPkTimer(msg)
					if inPkTimer {
						return
					}
					for _, fn := range withFns {
						out, err := fn(msg, pluginMap)
						if err != nil {
							return
						}
						// 前面处理过了, 那就不需要再处理
						if out != nil {
							if err = r.SendMessage(*out); err != nil {
								logger.Errorf("发送消息失败 %s\n", err)
							}
							return
						}
					}
				}
			}
			f()
		}
	}()

	r.Run(inCh)
}
