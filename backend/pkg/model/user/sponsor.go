package user

import basemodel "github.com/guojia99/cubing-pro/backend/pkg/model/base"

// Organizers 主办团队
type Organizers struct {
	basemodel.Model

	Name         string `gorm:"unique;not null;column:name"`
	Introduction string `gorm:"column:introduction"`
	Email        string `gorm:"column:email"`    // 邮箱
	QQGroup      string `gorm:"column:qq_group"` // QQ群

	LeaderID string // 组长
}

type AssOrganizerUsers struct {
	basemodel.Model

	OrganizersId uint `gorm:"index:,unique,composite:AssOrganizerUsers"`
	UserId       uint `gorm:"index:,unique,composite:AssOrganizerUsers"`
}
