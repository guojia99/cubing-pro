package user

import basemodel "github.com/guojia99/cubing-pro/backend/pkg/model/base"

// Organizers 主办团队
type Organizers struct {
	basemodel.Model

	Name         string `gorm:"unique;not null;column:name"`
	Introduction string `gorm:"column:introduction"`
	Email        string `gorm:"column:email"` // 邮箱

	LeaderID   string  // 组长
	Organizers []*User `gorm:"many2many:user_organizers"`
}
