package types

type InMessage struct {
	QQ      int64  `json:"QQ"`
	Name    string `json:"Name"`
	Message string `json:"Message"`
	GroupID int64  `json:"group_id"` // 群号
}

func (i InMessage) NewOutMessage(message string) *OutMessage {
	return &OutMessage{
		GroupID: i.GroupID,
		Message: message,
	}
}

type OutMessage struct {
	GroupID int64  `json:"GroupId"`
	Message string `json:"Message"`
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
