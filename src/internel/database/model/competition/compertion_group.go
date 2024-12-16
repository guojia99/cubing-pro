package competition

import basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"

type CompetitionGroup struct {
	basemodel.Model

	Name         string `gorm:"column:name" json:"name"`
	OrganizersID uint   `gorm:"column:orgId;null" json:"OrganizersID,omitempty"` // 主办团队

	QQGroups     string `gorm:"column:qq_groups" json:"qq_groups"`
	QQGroupUid   string `gorm:"column:qq_group_uid" json:"qq_group_uid"`
	WechatGroups string `gorm:"column:wechat_groups" json:"wechat_groups"`
}
