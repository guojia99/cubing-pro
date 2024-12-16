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

type MessageData struct {
	Text string `json:"text,omitempty"`
	QQ   int64  `json:"qq,omitempty"`
	File string `json:"file,omitempty"`
	Type string `json:"type,omitempty"` // flash 闪图， show 秀图 可以不用

	//0	正常图片
	//1	表情包, 在客户端会被分类到表情包图片并缩放显示
	//2	热图
	//3	斗图
	//4	智图?
	//7	贴图
	//8	自拍
	//9	贴图广告?
	//13 热搜图
	SubType int `json:"subType,omitempty"`
}

type Message struct {
	Data MessageData `json:"data"`
	Type string      `json:"type"`
}

type CqInMessage struct {
	MessageType MessageType `json:"message_type"`
	SelfID      int64       `json:"self_id"` // 自己的ID
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

type CQSendMessage struct {
	GroupId int64 `json:"group_id"`
	//QQId       int    `json:"-"`
	//Image      string `json:"-"`
	Message    []Message `json:"message"`
	AutoEscape bool      `json:"auto_escape"`
}
