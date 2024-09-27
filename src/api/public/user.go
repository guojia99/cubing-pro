package public

import (
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
)

type User struct {
	ID           uint   `json:"id"`
	Name         string `json:"Name"`                   // 名称
	EnName       string `json:"EnName"`                 // 英文名称
	CubeID       string `json:"CubeID"`                 // CubeID
	DelegateName string `json:"DelegateName,omitempty"` // 代表称呼: 高级代表\代表\实习代表...
	Avatar       string `json:"Avatar"`                 // 头像
	Level        uint   `json:"Level"`                  // 等级
	WcaID        string `json:"WcaID"`                  // wca ID
	Sign         string `json:"Sign"`                   // 签名
	Ban          bool   `json:"Ban"`                    //  封禁
}

func UserToUser(u user.User) User {
	return User{
		ID:           u.ID,
		Name:         u.Name,
		EnName:       u.EnName,
		CubeID:       u.CubeID,
		DelegateName: u.DelegateName,
		Avatar:       u.Avatar,
		Level:        u.Level,
		WcaID:        u.WcaID,
		Sign:         u.Sign,
		Ban:          u.Ban,
	}
}
