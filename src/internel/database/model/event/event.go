package event

import (
	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
)

type Event struct {
	basemodel.StringIDModel

	Name       string    `gorm:"column:name" json:"name"`              // 项目名
	Class      string    `gorm:"column:class" json:"class"`            // 分类
	IsComp     bool      `gorm:"column:is_comp" json:"isComp"`         // 比赛项目
	Icon       string    `gorm:"column:icon" json:"icon"`              // Icon
	IconBase64 string    `gorm:"column:icon_base64" json:"iconBase64"` // Icon base64
	IsWCA      bool      `gorm:"column:is_wca" json:"isWCA"`           // WCA项目
	RouteType  RouteType `gorm:"column:route_typ" json:"routeType"`    // 轮次
}
