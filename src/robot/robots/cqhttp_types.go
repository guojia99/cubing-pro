package robots

type MessageType = string

const (
	MessageTypePrivate = "private" // 私聊消息
	MessageTypeGroup   = "group"   // 群消息
)

type SubType = string

const (
	SubTypeFriend    = "friend"     // 好友
	SubTypeNormal    = "normal"     // 群聊
	SubTypeAnonymous = "anonymous"  // 	匿名
	SubTypeGroupSelf = "group_self" //	群中自身发送
	SubTypeGroup     = "group"      //群临时会话
	SubTypeNotice    = "notice"     //系统提示
)

type Sender struct {
	UserId   int64  `json:"user_id"` // QQ
	NickName string `json:"nickname"`
	Sex      string `json:"sex"` //  male 或 female 或 unknown
	Age      int32  `json:"age"`
	Card     string `json:"card"`
	Area     string `json:"area"`
	Level    string `json:"level"`
	Role     string `json:"role"`
	Title    string `json:"title"`
}

type Message struct {
	Data struct {
		Text string `json:"text"`
	}
	Type string `json:"type"`
}

type CqInMessage struct {
	MessageType MessageType `json:"message_type"`
	SubType     SubType     `json:"sub_type"`
	MessageId   int32       `json:"message_id"`
	UserId      int64       `json:"user_id"` // QQ
	Message     []Message   `json:"message"` // 消息
	RawMessage  string      `json:"raw_message"`
	Font        int         `json:"font"`     // 字体
	GroupID     int64       `json:"group_id"` // 群号
	Sender      Sender      `json:"sender"`

	MetaEventType string `json:"meta_event_type"`
}
