package types

import (
	"fmt"

	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type InMessage struct {
	QQ      int64       `json:"QQ"`
	QQBot   string      `json:"QQBot"`
	Name    string      `json:"Name"`
	Message string      `json:"Message"`
	GroupID interface{} `json:"GroupID"` // 群号
	MsgID   string      `json:"MsgID"`   // 消息ID
}

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
