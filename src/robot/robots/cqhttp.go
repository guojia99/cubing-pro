package robots

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/robot/types"
	"net/http"
	"strings"
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
	fmt.Println(err)
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

	msg := types.InMessage{
		QQ:      r.Sender.UserId,
		Name:    r.Sender.NickName,
		Message: "",
		GroupID: r.GroupID,
	}

	for _, v := range r.Message {
		if v.Type == "text" {
			msg.Message += v.Data.Text
		}
	}

	msg.Message = strings.TrimPrefix(msg.Message, " ")

	select {
	case <-ctx.Done():
		return
	case c.ch <- msg:
		return
	}
}

type CQSendMessage struct {
	GroupId    int64  `json:"group_id"`
	QQId       int    `json:"-"`
	Image      string `json:"-"`
	Message    string `json:"message"`
	AutoEscape bool   `json:"auto_escape"`
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
	_, err := utils.HTTPRequest(
		"POST", fmt.Sprintf("%s/send_group_msg", c.cfg.Address), nil, nil, CQSendMessage{
			GroupId:    out.GroupID,
			Message:    out.Message,
			AutoEscape: false,
		},
	)
	return err
}
