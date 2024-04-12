package post

import basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"

type Notification struct {
	basemodel.Model

	Title          string `gorm:"column:title"`             // 标题
	Short          string `gorm:"column:short"`             // 简短
	Type           string `gorm:"column:type"`              // 通知类型
	Top            bool   `gorm:"column:top"`               // 是否置顶
	Content        string `gorm:"column:content"`           // markdown
	CreateBy       string `gorm:"column:create_by"`         // 创建人
	CreateByUserID string `gorm:"column:create_by_user_id"` // 创建人ID
	Remark         string `gorm:"column:remark"`            // 备注
}
