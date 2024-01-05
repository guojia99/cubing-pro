package user

import basemodel "github.com/guojia99/cubing-pro/model/base"

// Organizers 主办
type Organizers struct {
	basemodel.Model

	Name         string `gorm:"unique;not null;column:name"`
	Introduction string `gorm:"column:introduction"`
	Email        string `gorm:"column:email"` // 邮箱

	LeaderID   uint // 组长
	Organizers []User
}
