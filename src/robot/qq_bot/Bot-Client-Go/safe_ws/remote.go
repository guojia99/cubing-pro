package safe_ws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/fanliao/go-promise"
	"github.com/gorilla/websocket"
	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/dto"
	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/onebot"
	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/openapi"
	log "github.com/sirupsen/logrus"
)

var Bots = make(map[string]*Bot)

type Bot struct {
	QQ        uint64
	AppId     string
	Token     string
	AppSecret string
	Openapi   openapi.OpenAPI

	mux           sync.RWMutex
	WaitingFrames map[string]*promise.Promise

	Payload *dto.WSPayload
}

var (
	FirstStart bool = true
)

func ConnectUniversal(appid, serverUrl string) {
	for {
		header := http.Header{}
		header.Add("x-bot-self-id", appid)
		conn, _, err := websocket.DefaultDialer.Dial(serverUrl, header)
		fmt.Println(err)
		if err != nil {
			log.Warnf("连接Websocket服务器 %s 错误，5秒后重连", serverUrl)
			time.Sleep(5 * time.Second)
			continue
		} else {
			log.Infof("连接Websocket服务器成功 %s", serverUrl)
			closeChan := make(chan int, 1)
			safeWs := NewSafeWebSocket(conn, func(ws *SafeWebSocket, messageType int, data []byte) {
				rm := onebot.Frame{}
				json.Unmarshal(data, &rm)
				fmt.Println(string(data))
				if FirstStart {
					NewBot(rm.BotId, rm.Payload, rm.Data)
					FirstStart = false
				}
				NewBot(rm.BotId, rm.Payload, rm.Data)
			}, func() {
				defer func() {
					_ = recover() // 可能多次触发
				}()
				closeChan <- 1
			})
			SafeGo(func() {
				if err := safeWs.Send(websocket.PingMessage, []byte("ping")); err != nil {
				}
				time.Sleep(5 * time.Second)
			})
			<-closeChan
			close(closeChan)
			log.Warnf("Websocket 服务器 %s 已断开，5秒后重连", serverUrl)
			time.Sleep(5 * time.Second)
		}
	}
}

func ConnectUniversalWithSecret(appid, secret, serverUrl string) {
	for {
		header := http.Header{}
		header.Add("x-bot-self-id", appid)
		header.Add("x-bot-secret", secret)
		conn, _, err := websocket.DefaultDialer.Dial(serverUrl, header)
		fmt.Println(err)
		if err != nil {
			log.Warnf("连接Websocket服务器 %s 错误，5秒后重连", serverUrl)
			time.Sleep(5 * time.Second)
			continue
		} else {
			log.Infof("连接Websocket服务器成功 %s", serverUrl)
			closeChan := make(chan int, 1)
			safeWs := NewSafeWebSocket(conn, func(ws *SafeWebSocket, messageType int, data []byte) {
				rm := onebot.Frame{}
				json.Unmarshal(data, &rm)
				fmt.Println(string(data))
				if FirstStart {
					NewBot(rm.BotId, rm.Payload, rm.Data)
					FirstStart = false
				}
				NewBot(rm.BotId, rm.Payload, rm.Data)
			}, func() {
				defer func() {
					_ = recover() // 可能多次触发
				}()
				closeChan <- 1
			})
			SafeGo(func() {
				if err := safeWs.Send(websocket.PingMessage, []byte("ping")); err != nil {
				}
				time.Sleep(5 * time.Second)
			})
			<-closeChan
			close(closeChan)
			log.Warnf("Websocket 服务器 %s 已断开，5秒后重连", serverUrl)
			time.Sleep(5 * time.Second)
		}
	}
}

func NewBot(appId string, p *dto.WSPayload, m []byte) *Bot {
	ibot, ok := Bots[appId]
	if ok {
		ibot.ParseWHData(appId, p, m)
	}
	bot := &Bot{
		AppId:   appId,
		Payload: p,
	}
	Bots[bot.AppId] = bot
	return bot
}

func (bot *Bot) ParseWHData(h string, p *dto.WSPayload, message []byte) {
	if p.Type == dto.EventGroupATMessageCreate {
		gm := &dto.WSGroupATMessageData{}
		err := json.Unmarshal(message, gm)
		if err == nil {
			GroupAtMessageEventHandler(h, p, gm)
		}
	}
	if p.Type == dto.EventGroupAddRobot {
		fmt.Println(p.Type)
		gar := &dto.WSGroupAddRobotData{}
		err := json.Unmarshal(message, gar)
		fmt.Println(err)
		if err == nil {
			GroupAddRobotEventHandler(h, p, gar)
		}
	}
	if p.Type == dto.EventGroupDelRobot {
		gdr := &dto.WSGroupDelRobotData{}
		err := json.Unmarshal(message, gdr)
		if err == nil {
			GroupDelRobotEventHandler(h, p, gdr)
		}
	}
	if p.Type == dto.EventGroupMsgReceive {
		gmr := &dto.WSGroupMsgReceiveData{}
		err := json.Unmarshal(message, gmr)
		if err == nil {
			GroupMsgReceiveEventHandler(h, p, gmr)
		}
	}
	if p.Type == dto.EventGroupMsgReject {
		gmr := &dto.WSGroupMsgRejectData{}
		err := json.Unmarshal(message, gmr)
		if err == nil {
			GroupMsgRejectEventHandler(h, p, gmr)
		}
	}
	if p.Type == dto.EventC2CMessageCreate {
		cmc := &dto.WSC2CMessageData{}
		err := json.Unmarshal(message, cmc)
		if err == nil {
			C2CMessageEventHandler(h, p, cmc)
		}
	}
	if p.Type == dto.EventC2CMsgReceive {
		fmr := &dto.WSFriendMsgReveiceData{}
		err := json.Unmarshal(message, fmr)
		if err == nil {
			C2CMsgReceiveHandler(h, p, fmr)
		}
	}
	if p.Type == dto.EventC2CMsgReject {
		fmr := &dto.WSFriendMsgRejectData{}
		err := json.Unmarshal(message, fmr)
		if err == nil {
			C2CMsgRejectHandler(h, p, fmr)
		}
	}
	if p.Type == dto.EventFriendAdd {
		fad := &dto.WSFriendAddData{}
		err := json.Unmarshal(message, fad)
		if err == nil {
			FriendAddEventHandler(h, p, fad)
		}
	}
	if p.Type == dto.EventFriendDel {
		fad := &dto.WSFriendDelData{}
		err := json.Unmarshal(message, fad)
		if err == nil {
			FriendDelEventHandler(h, p, fad)
		}
	}
	if p.Type == dto.EventAtMessageCreate {
		am := &dto.WSATMessageData{}
		err := json.Unmarshal(message, am)
		if err == nil {
			ATMessageEventHandler(h, p, am)
		}
	}
	if p.Type == dto.EventMessageCreate {
		m := &dto.WSMessageData{}
		err := json.Unmarshal(message, m)
		if err == nil {
			MessageEventHandler(h, p, m)
		}
	}
	if p.Type == dto.EventInteractionCreate {
		i := &dto.WSInteractionData{}
		err := json.Unmarshal(message, i)
		if err == nil {
			i.ID = p.ID
			InteractionEventHandler(h, p, i)
		}
	}
	if p.Type == dto.EventDirectMessageCreate {
		i := &dto.WSDirectMessageData{}
		err := json.Unmarshal(message, i)
		if err == nil {
			DirectMessageEventHandler(h, p, i)
		}
	}
	if p.Type == dto.EventMessageReactionAdd || p.Type == dto.EventMessageReactionRemove {
		mr := &dto.WSMessageReactionData{}
		err := json.Unmarshal(message, mr)
		if err == nil {
			MessageReactionEventHandler(h, p, mr)
		}
	}
	if p.Type == dto.EventMessageAuditPass || p.Type == dto.EventMessageAuditReject {
		mr := &dto.WSMessageAuditData{}
		err := json.Unmarshal(message, mr)
		if err == nil {
			MessageAuditEventHandler(h, p, mr)
		}
	}
	if p.Type == dto.EventForumThreadCreate || p.Type == dto.EventForumPostCreate || p.Type == dto.EventForumReplyCreate || p.Type == dto.EventForumThreadUpdate || p.Type == dto.EventForumPostDelete || p.Type == dto.EventForumThreadDelete || p.Type == dto.EventForumReplyDelete {
		ft := &dto.WSForumAuditData{}
		err := json.Unmarshal(message, ft)
		if err == nil {
			ForumAuditEventHandler(h, p, ft)
		}
	}
	if p.Type == dto.EventGuildCreate || p.Type == dto.EventGuildUpdate || p.Type == dto.EventGuildDelete {
		g := &dto.WSGuildData{}
		err := json.Unmarshal(message, g)
		if err == nil {
			GuildEventHandler(h, p, g)
		}
	}
	if p.Type == dto.EventChannelCreate || p.Type == dto.EventChannelUpdate || p.Type == dto.EventChannelDelete {
		c := &dto.WSChannelData{}
		err := json.Unmarshal(message, c)
		if err == nil {
			ChannelEventHandler(h, p, c)
		}
	}
	if p.Type == dto.EventGuildMemberAdd || p.Type == dto.EventGuildMemberUpdate || p.Type == dto.EventGuildMemberRemove {
		gm := &dto.WSGuildMemberData{}
		err := json.Unmarshal(message, gm)
		if err == nil {
			GuildMemberEventHandler(h, p, gm)
		}
	}
}
