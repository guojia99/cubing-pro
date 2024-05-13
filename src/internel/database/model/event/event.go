package event

import (
	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
)

type Event struct {
	basemodel.StringIDModel `table:"-"`

	Name          string    `gorm:"column:name" json:"name"`                        // 项目名
	OtherNames    string    `gorm:"column:other_name" json:"otherNames"`            // 其他名称
	Class         string    `gorm:"column:class" json:"class"`                      // 分类
	IsComp        bool      `gorm:"column:is_comp" json:"isComp"`                   // 比赛项目
	Icon          string    `gorm:"column:icon" json:"icon" table:"-"`              // Icon
	IconBase64    string    `gorm:"column:icon_base64" json:"iconBase64" table:"-"` // Icon base64
	IsWCA         bool      `gorm:"column:is_wca" json:"isWCA"`                     // WCA项目
	BaseRouteType RouteType `gorm:"column:base_route_typ" json:"base_route_typ"`    // 默认轮次
}
