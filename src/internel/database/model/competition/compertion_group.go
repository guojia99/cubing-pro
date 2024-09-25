package competition

import basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"

type CompertionGroup struct {
	basemodel.Model

	Name string `gorm:"column:name" json:"name"`

	QQGroups     string `gorm:"column:qq_groups" json:"qq_groups"`
	QQGroupUid   string `gorm:"column:qq_group_uid" json:"qq_group_uid"`
	WechatGroups string `gorm:"column:wechat_groups" json:"wechat_groups"`
}
