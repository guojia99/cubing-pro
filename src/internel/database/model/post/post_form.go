package post

import basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"

// Forum 板块表
type Forum struct {
	basemodel.Model

	Name   string `gorm:"column:name"`
	Remark string `gorm:"column:remark"` // 备注
}

type TopicStatus int

const (
	TopicStatusRelease       TopicStatus = iota + 1 //发布
	TopicStatusUnpublished                          //未发布
	TopicStatusBan                                  //封禁
	TopicStatusPendingReview                        //待审核
)

// Topic 主题帖
type Topic struct {
	basemodel.Model

	Fid            uint   `gorm:"column:fid"`               // 板块id
	CreateBy       string `gorm:"column:create_by"`         // 创建人
	CreateByUserID uint   `gorm:"column:create_by_user_id"` // 创建人ID

	Status     TopicStatus `gorm:"column:status"`            // 发布状态
	Title      string      `gorm:"column:title"`             // 标题
	Short      string      `gorm:"column:short"`             // 简短说明
	Content    string      `gorm:"column:content"`           // md
	Tags       string      `gorm:"column:tags;index:,"`      // tags
	Type       string      `gorm:"column:type"`              // 类型
	TopImage   string      `gorm:"column:top_img"`           // 头图
	IsOriginal bool        `gorm:"column:is_original"`       // 是否原创
	Original   string      `gorm:"column:original"`          // 原创
	KeyWords   string      `gorm:"column:key_words;index:,"` // 关键词
	Ip         string      `gorm:"column:ip"`                // ip地址
}

type AssTopicLike struct {
	basemodel.Model

	Uid uint `gorm:"index:,unique,composite:AssTopicLike"`
	Tid uint `gorm:"index:,unique,composite:AssTopicLike"`
}

// Posts 回复内容
type Posts struct {
	basemodel.Model

	Tid      uint `gorm:"column:tid"` // 帖子id
	Uid      uint `gorm:"column:uid"` // 用户id
	ReplyPid uint `gorm:"reply_pid"`  // 回复的pid

	UserName string `gorm:"column:user_name"` // 用户名
	Content  string `gorm:"column:content"`   // 回复内容
	IP       string `gorm:"column:ip"`        // ip地址
}

type AssPostsLike struct {
	basemodel.Model

	Uid uint `gorm:"index:,unique,composite:AssPostsLike"`
	Pid uint `gorm:"index:,unique,composite:AssPostsLike"`
}
