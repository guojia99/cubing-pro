package robots

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	//bot "github.com/2mf8/Better-Bot-Go"
	//"github.com/2mf8/Better-Bot-Go/dto"
	//"github.com/2mf8/Better-Bot-Go/openapi"
	//v1 "github.com/2mf8/Better-Bot-Go/openapi/v1"
	//"github.com/2mf8/Better-Bot-Go/token"
	//"github.com/2mf8/Better-Bot-Go/webhook"
	//"github.com/2mf8/Bot-Client-Go/safe_ws"

	bot "github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go"
	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/dto"
	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/openapi"
	v1 "github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/openapi/v1"
	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/token"
	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/webhook"
	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Bot-Client-Go/safe_ws"

	"github.com/donnie4w/go-logger/logger"
	"github.com/gin-gonic/gin"

	"github.com/guojia99/cubing-pro/src/internel/configs"
	"github.com/guojia99/cubing-pro/src/robot/types"
	"github.com/patrickmn/go-cache"
)

type QQBot struct {
	Api openapi.OpenAPI

	cfg *configs.QQBotConfig
	ctx context.Context
	ch  chan<- types.InMessage

	serverGin *gin.Engine

	ReqCache *cache.Cache
}

func NewQQBot(cfg *configs.QQBotConfig, ctx context.Context) *QQBot {
	return &QQBot{
		cfg:      cfg,
		ctx:      ctx,
		ReqCache: cache.New(100*time.Minute, 100*time.Minute),
	}
}

func (q *QQBot) Prefix() string {
	return ""
}

func (q *QQBot) runServer() {
	webhook.AllSetting = &webhook.Setting{
		Apps: map[string]*webhook.App{
			"qq-robot": {
				QQ:        q.cfg.QQ,
				AppId:     q.cfg.AppId,
				Token:     q.cfg.Token,
				AppSecret: q.cfg.AppSecret,
				IsSandBox: false,
				WSSAddr:   q.cfg.WSSAddr,
			},
		},
		Port:     q.cfg.Server.Port,
		CertFile: q.cfg.Server.CertFile,
		CertKey:  q.cfg.Server.CertKey,
	}
	webhook.InitLog()
	webhook.InitGin(q.cfg.Server.IsOpen)
}

func (q *QQBot) updateToken() {
	atr := v1.GetAccessToken(fmt.Sprintf("%v", q.cfg.AppId), q.cfg.AppSecret)
	iat, _ := strconv.Atoi(atr.ExpiresIn)
	aei := time.Now().Unix() + int64(iat)
	tk := token.BotToken(q.cfg.AppId, atr.AccessToken, string(token.TypeQQBot))

	if q.cfg.IsSandBox {
		q.Api = bot.NewSandboxOpenAPI(tk).WithTimeout(30 * time.Second)
	} else {
		q.Api = bot.NewOpenAPI(tk).WithTimeout(30 * time.Second)
	}
	bot.AuthAcessAdd(fmt.Sprintf("%v", q.cfg.AppId), &bot.AccessToken{
		AccessToken: atr.AccessToken,
		ExpiresIn:   aei,
		Api:         q.Api,
		AppSecret:   q.cfg.AppSecret,
		IsSandBox:   q.cfg.IsSandBox,
		Appid:       q.cfg.AppId,
	})

}

func (q *QQBot) Run(ch chan<- types.InMessage) {
	//safe_ws.InitLog()
	// 服务端
	go q.runServer()

	// 客户端
	q.updateToken()

	q.ch = ch
	safe_ws.GroupAtMessageEventHandler = q.messageAtEventHandler

	if q.cfg.IsOpen {
		go safe_ws.ConnectUniversalWithSecret(fmt.Sprintf("%v", q.cfg.AppId), q.cfg.AppSecret, q.cfg.WSSAddr)
	} else {
		go safe_ws.ConnectUniversal(fmt.Sprintf("%v", q.cfg.AppId), q.cfg.WSSAddr)
	}

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
	var msgReq uint = 0
	req, ok := q.ReqCache.Get(out.MsgID)
	if ok {
		msgReq = req.(uint)
	}
	msgReq += 1
	q.ReqCache.Set(out.MsgID, msgReq, time.Minute*100)

	//TODO implement me
	newMsg := &dto.GroupMessageToCreate{
		Content: "\n" + strings.Join(out.Message, ""),
		MsgID:   out.MsgID,
		MsgReq:  msgReq,
		MsgSeq:  msgReq,
		MsgType: dto.C2CMsgTypeText,
	}

	if len(out.Images) == 1 {
		data, err := os.ReadFile(out.Images[0])
		if err != nil {
			return err
		}

		// 发送图片
		resp, err := bot.SendApi(fmt.Sprintf("%v", q.cfg.AppId)).PostGroupRichMediaMessage(q.ctx, out.GroupID.(string),
			&dto.GroupRichMediaMessageToCreate{
				FileType:   1,
				FileData:   data,
				SrvSendMsg: false,
			},
		)
		if err != nil {
			return err
		}

		// 重新组合图文格式
		newMsg.MsgType = dto.C2CMsgTypeMedia
		newMsg.Media = &dto.FileInfo{
			FileInfo: resp.FileInfo,
		}
	}

	_, err := bot.SendApi(fmt.Sprintf("%v", q.cfg.AppId)).PostGroupMessage(q.ctx, out.GroupID.(string), newMsg)
	return err
}
