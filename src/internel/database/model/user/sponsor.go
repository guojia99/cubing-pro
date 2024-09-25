package user

import (
	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	jsoniter "github.com/json-iterator/go"
)

type OrganizersStatus = string

const (
	NotUse              OrganizersStatus = "NotUse"
	Expired             OrganizersStatus = "Expired"             // 过期
	Using               OrganizersStatus = "Using"               // 使用中
	Applying            OrganizersStatus = "Applying"            // 申请中
	RejectApply         OrganizersStatus = "RejectApply"         // 驳回申请
	UnderAppeal         OrganizersStatus = "UnderAppeal"         // 申诉中
	RejectAppeal        OrganizersStatus = "RejectAppeal"        // 驳回申诉
	Disable             OrganizersStatus = "Disable"             // 禁用
	PermanentlyDisabled OrganizersStatus = "PermanentlyDisabled" // 永久禁用
	Disband             OrganizersStatus = "Disband"             // 解散 无法使用
)

// Organizers 主办团队
type Organizers struct {
	basemodel.Model

	Name         string `gorm:"unique;not null;column:name"` // 名
	Introduction string `gorm:"column:introduction"`         // 介绍 md
	Email        string `gorm:"column:email"`                // 邮箱

	LeaderID          string           `gorm:"column:leaderId"`      // 组长 cubeID
	AssOrganizerUsers string           `gorm:"column:ass_org_users"` // 成员列表
	Status            OrganizersStatus `gorm:"column:status"`        // 状态

	LeaderRemark string `gorm:"column:leader_remark"` // 组长备注
	AdminMessage string `gorm:"column:admin_msg"`     // 管理员留言
}

func (o *Organizers) CanUse() bool {
	switch o.Status {
	case Using:
		return true
	}
	return false
}

func (o *Organizers) SetUsersCubingID(u []string) {
	var old []string
	_ = jsoniter.UnmarshalFromString(o.AssOrganizerUsers, &old)
	newL := utils.Merge[string](old, u)
	o.AssOrganizerUsers = utils.ToJSON(newL)
	return
}

func (o *Organizers) DeleteUserID(u []string) {
	var old []string
	_ = jsoniter.UnmarshalFromString(o.AssOrganizerUsers, &old)
	newL := utils.Delete[string](old, u)
	o.AssOrganizerUsers = utils.ToJSON(newL)
	return
}

func (o *Organizers) Users() []string {
	var out []string
	_ = jsoniter.UnmarshalFromString(o.AssOrganizerUsers, &out)
	return append([]string{o.LeaderID}, out...)
}
