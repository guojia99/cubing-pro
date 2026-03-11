package user

import (
	"time"

	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
)

// OAuthState WCA OAuth 授权流程中的 state 持久化存储，用于防 CSRF
// 服务器重启后仍可校验，替代内存 cache
type OAuthState struct {
	basemodel.Model

	Nonce     string    `gorm:"column:nonce;uniqueIndex;size:64;not null" json:"-"` // 随机 nonce
	Redirect  string    `gorm:"column:redirect;" json:"-"`                          // 登录成功后跳转地址
	ExpiresAt time.Time `gorm:"column:expires_at;not null" json:"-"`                // 过期时间
}

func (OAuthState) TableName() string {
	return "oauth_states"
}
