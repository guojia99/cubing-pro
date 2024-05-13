package public

import (
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
)

type User struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`         // 名称
	EnName       string `json:"enName"`       // 英文名称
	CubeID       string `json:"cubeID"`       // CubeID
	DelegateName string `json:"delegateName"` // 代表称呼: 高级代表\代表\实习代表...
	Avatar       string `json:"avatar"`       // 头像
}

func UserToUser(u user.User) User {
	return User{
		ID:           u.ID,
		Name:         u.Name,
		EnName:       u.EnName,
		CubeID:       u.CubeID,
		DelegateName: u.DelegateName,
		Avatar:       u.Avatar,
	}
}
