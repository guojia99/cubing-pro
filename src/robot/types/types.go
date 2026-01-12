package types

import (
	"fmt"
	"strings"

	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type Token string

//func (t Token) ToNumber() int {
//	out, _ := strconv.Atoi(string(t))
//	return out
//}
//
//func (t Token) ToFloat() string {
//
//}

type InMessage struct {
	QQ      int64       `json:"QQ,omitempty"`
	QQBot   string      `json:"QQBot,omitempty"`
	Name    string      `json:"Name,omitempty"`
	Message string      `json:"Message,omitempty"`
	Tokens  []Token     `json:"Tokens,omitempty"`
	GroupID interface{} `json:"GroupID,omitempty"` // 群号
	MsgID   string      `json:"MsgID,omitempty"`   // 消息ID
}

func (i InMessage) GroupIDStr() string { return i.GroupID.(string) }

func (i InMessage) NewOutMessage(message ...string) *OutMessage {
	var msg []string
	for _, m := range message {
		msg = append(msg, utils.RemoveEmptyLines(m))
	}

	out := &OutMessage{
		GroupID: i.GroupID,
		Message: msg,
		MsgID:   i.MsgID,
	}

	return out
}

func (i InMessage) NewOutMessagef(format string, a ...any) *OutMessage {
	msg := fmt.Sprintf(format, a...)
	return i.NewOutMessage(msg)
}

func (i InMessage) NewOutMessageWithImage(msg string, images ...string) *OutMessage {
	out := i.NewOutMessage(msg)
	out.Images = images
	return out
}

type OutMessage struct {
	// cqHttp
	GroupID interface{} `json:"GroupId"`
	MsgID   string      `json:"MsgID"`
	Message []string    `json:"Message"`
	Images  []string    `json:"Images"`
}

func (o *OutMessage) AddMessagef(format string, a ...any) *OutMessage {
	o.Message = append(o.Message, fmt.Sprintf(format, a...))
	return o
}
func (o *OutMessage) AddMessages(msg ...string) *OutMessage {
	o.Message = append(o.Message, msg...)
	return o
}

type Plugin interface {
	ID() []string
	Help() string
	Do(message InMessage) (*OutMessage, error)
}

type Robot interface {
	Prefix() string
	Run(ch chan<- InMessage)
	SendMessage(out OutMessage) error
}

func RemoveID(message string, id []string) string {
	for _, i := range id {
		message = strings.TrimLeft(message, i)
	}
	return message
}
