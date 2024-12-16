package types

import "github.com/guojia99/cubing-pro/src/internel/utils"

type InMessage struct {
	QQ      int64  `json:"QQ"`
	Name    string `json:"Name"`
	Message string `json:"Message"`
	GroupID int64  `json:"group_id"` // 群号
}

func (i InMessage) NewOutMessage(message ...string) *OutMessage {
	var msg []string
	for _, m := range message {
		msg = append(msg, utils.RemoveEmptyLines(m))
	}

	return &OutMessage{
		GroupID: i.GroupID,
		Message: msg,
	}
}

func (i InMessage) NewOutMessageWithImage(msg string, images ...string) *OutMessage {
	return &OutMessage{
		GroupID: i.GroupID,
		Message: []string{msg},
		Images:  images,
	}
}

type OutMessage struct {
	GroupID int64    `json:"GroupId"`
	Message []string `json:"Message"`
	Images  []string `json:"Images"`
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
