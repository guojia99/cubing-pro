package user

import (
	"time"

	"gorm.io/gorm"
)

const (
	RegisterWithEmail            = "reg_email"
	RetrievePasswordWithEmail    = "ret_email"
	RetrievePasswordWithAdminKey = "admin_key" // 授权码找回密码
	// RetrievePasswordWithPhone    = "ret_phone"
)

type CheckCode struct {
	gorm.Model

	Type    string    `gorm:"column:typ"`     // 验证类型
	UserID  uint      `gorm:"column:user_id"` // 用户ID
	Email   string    `gorm:"column:email"`   // 邮箱
	Use     bool      `gorm:"column:use"`     // 是否已经使用
	Key     string    `gorm:"column:key"`     // 验证key
	Code    string    `gorm:"column:code"`    // 验证码
	Timeout time.Time `gorm:"column:timeout"` // 验证码超时时间
}
