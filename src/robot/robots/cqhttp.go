package robots

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/donnie4w/go-logger/logger"
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/robot/types"
)

type CqHttps struct {
	api *gin.Engine

	ch  chan<- types.InMessage
	cfg *svc.CQHttpBot
}

func NewCqHttps(cfg *svc.CQHttpBot) *CqHttps {
	return &CqHttps{cfg: cfg, api: gin.Default()}
}

func (c *CqHttps) Run(ch chan<- types.InMessage) {
	c.ch = ch
	c.api.NoRoute(c.route)
	err := c.api.Run(fmt.Sprintf("127.0.0.1:%d", c.cfg.Post))
	logger.Error(err)
}

func (c *CqHttps) Prefix() string {
	return c.cfg.Prefix
}

func (c *CqHttps) route(ctx *gin.Context) {
	var r CqInMessage
	_ = ctx.Bind(&r)
	ctx.JSON(http.StatusOK, gin.H{})

	if r.MetaEventType == "heartbeat" {
		return
	}
	if r.MessageType != MessageTypeGroup {
		return
	}
	// 判断是不是艾特自己
	if r.SelfID == r.UserId {
		return
	}

	msg := types.InMessage{
		QQ:      r.Sender.UserId,
		Name:    r.Sender.NickName,
		Message: "",
		GroupID: r.GroupID,
	}

	for _, v := range r.Message {
		switch v.Type {
		case "text":
			msg.Message += v.Data.Text
		case "at":
			// 解析raw message 是否有艾特自己
			// [CQ:at,qq=2854216320]
			atMe := fmt.Sprintf("[CQ:at,qq=%d]", r.SelfID)
			if v.Data.QQ == r.SelfID ||
				(strings.Contains(r.RawMessage, "[CQ:at,qq=") && !strings.Contains(r.RawMessage, atMe)) {
				return
			}
		}
	}

	data, _ := json.Marshal(r)
	logger.Infof("[Robot] [CQ] 处理消息 %s", string(data))

	msg.Message = strings.TrimPrefix(msg.Message, " ")

	select {
	case <-ctx.Done():
		return
	case c.ch <- msg:
		return
	}
}

func (c *CqHttps) SendMessage(out types.OutMessage) error {
	//if message[len(message)-1] == '\n' {
	//	message = message[:len(message)-1]
	//}
	//
	//if qqId != 0 {
	//	message = fmt.Sprintf("[CQ:at,qq=%d]\n", qqId) + message
	//}
	//
	//if c.cfg.NotMessage {
	//	log.Printf("%s\n", message)
	//	return nil
	//}
	//
	//if imagePath != "" {
	//	message += fmt.Sprintf("\n[CQ:image,file=file:///%s]", imagePath)
	//}
	//
	//if out

	msg := CQSendMessage{
		GroupId:    out.GroupID.(int64),
		Message:    []Message{},
		AutoEscape: false,
	}

	for _, v := range out.Message {
		msg.Message = append(msg.Message, Message{
			Data: MessageData{
				Text: v,
			},
			Type: "text",
		})
	}

	for _, v := range out.Images {
		msg.Message = append(msg.Message, Message{
			Data: MessageData{
				File:    "file://" + v,
				Type:    "show",
				SubType: 0,
			},
			Type: "image",
		})
	}

	// todo 使用CQ码

	_, err := utils.HTTPRequest(
		"POST", fmt.Sprintf("%s/send_group_msg", c.cfg.Address), nil, nil, msg,
	)

	data, _ := json.Marshal(msg)
	logger.Infof("[Robot][CQ] 发送消息 %s", string(data))

	return err
}
