package post

import (
	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
)

type Notification struct {
	basemodel.Model

	Title          string `gorm:"column:title" json:"title"`                      // 标题
	Short          string `gorm:"column:short" json:"short"`                      // 简短
	Type           string `gorm:"column:type" json:"type"`                        // 通知类型
	Top            bool   `gorm:"column:top" json:"top"`                          // 是否置顶
	Fixed          bool   `gorm:"column:fixed" json:"fixed"`                      // 是否侧边
	Content        string `gorm:"column:content" json:"content,omitempty"`        // markdown
	CreateBy       string `gorm:"column:create_by" json:"createBy"`               // 创建人
	CreateByUserID uint   `gorm:"column:create_by_user_id" json:"createByUserID"` // 创建人ID
	Remark         string `gorm:"column:remark" json:"remark,omitempty"`          // 备注
}
