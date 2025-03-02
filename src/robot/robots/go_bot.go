package robots

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	bot "github.com/2mf8/Better-Bot-Go"
	"github.com/2mf8/Better-Bot-Go/dto"
	"github.com/2mf8/Better-Bot-Go/openapi"
	"github.com/2mf8/Better-Bot-Go/token"
	"github.com/2mf8/Bot-Client-Go/safe_ws"
	"github.com/donnie4w/go-logger/logger"
	"github.com/guojia99/cubing-pro/src/internel/configs"
	"github.com/guojia99/cubing-pro/src/robot/types"
)

type QQBot struct {
	Api openapi.OpenAPI

	cfg *configs.QQBotConfig
	ctx context.Context
	ch  chan<- types.InMessage
}

func NewQQBot(cfg *configs.QQBotConfig, ctx context.Context) *QQBot {
	return &QQBot{
		cfg: cfg,
		ctx: ctx,
	}
}

func (q *QQBot) Prefix() string {
	return ""
}

func (q *QQBot) Run(ch chan<- types.InMessage) {
	//safe_ws.InitLog()

	go safe_ws.ConnectUniversal(fmt.Sprintf("%v", q.cfg.AppId), q.cfg.WSSAddr)

	q.Api = bot.NewSandboxOpenAPI(
		token.BotToken(q.cfg.AppId, q.cfg.Token, string(token.TypeBot)),
	).WithTimeout(3 * time.Second)

	q.ch = ch
	safe_ws.GroupAtMessageEventHandler = q.messageAtEventHandler
	//safe_ws.GroupMessageEventHandler = q.messageEventHandler
	select {
	case <-q.ctx.Done():
		return
	}
}

//func (q *QQBot) messageEventHandler(appid string, event *dto.WSPayload, data *dto.WSGroupMessageData) error {
//	log.Info(data.Content, data.GroupId)
//	log.Infof("%+v\n", data)
//
//	log.Infof("%+v\n", data.Author.UserId)
//	log.Infof("%+v\n", data.Author.UserOpenId)
//	return nil
//}

func (q *QQBot) messageAtEventHandler(appid string, event *dto.WSPayload, data *dto.WSGroupATMessageData) error {

	msg := types.InMessage{
		QQ:      0,
		QQBot:   data.Author.UserOpenId,
		Name:    "",
		Message: data.Content,
		GroupID: data.GroupId,
		MsgID:   data.MsgId,
	}

	msg.Message = strings.TrimLeft(msg.Message, " ")

	select {
	case <-q.ctx.Done():
		return nil
	case q.ch <- msg:
		d, _ := json.Marshal(msg)
		logger.Infof("[Robot] [GoBot] 处理消息 %s", string(d))
		return nil
	}
}

func (q *QQBot) SendMessage(out types.OutMessage) error {
	//TODO implement me
	newMsg := &dto.GroupMessageToCreate{
		Content: "\n" + strings.Join(out.Message, ""),
		MsgID:   out.MsgID,
		MsgReq:  1,
		MsgType: 0,
	}

	if len(out.Images) == 1 {

		data, err := os.ReadFile(out.Images[0])
		if err != nil {
			return err
		}

		resp, err := q.Api.PostGroupRichMediaMessage(q.ctx, out.GroupID.(string),
			&dto.GroupRichMediaMessageToCreate{
				FileType:   1,
				FileData:   data,
				SrvSendMsg: false,
			},
		)
		if err != nil {
			return err
		}

		newMsg.MsgType = 7
		newMsg.Media = &dto.FileInfo{
			FileInfo: resp.FileInfo,
		}
	}

	_, err := q.Api.PostGroupMessage(q.ctx, out.GroupID.(string), newMsg)
	return err
}
