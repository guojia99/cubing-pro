package event

import (
	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
)

type Event struct {
	basemodel.StringIDModel `table:"-"`
	Idx                     int64     `gorm:"column:idx" json:"idx,omitempty"`                          // 项目排序
	Name                    string    `gorm:"column:name" json:"name,omitempty"`                        // 项目名
	OtherNames              string    `gorm:"column:other_name" json:"otherNames,omitempty"`            // 其他名称
	Class                   string    `gorm:"column:class" json:"class,omitempty"`                      // 分类
	IsComp                  bool      `gorm:"column:is_comp" json:"isComp,omitempty"`                   // 比赛项目
	Icon                    string    `gorm:"column:icon" json:"icon,omitempty" table:"-"`              // Icon
	IconBase64              string    `gorm:"column:icon_base64" json:"iconBase64,omitempty" table:"-"` // Icon base64
	IsWCA                   bool      `gorm:"column:is_wca" json:"isWCA,omitempty"`                     // WCA项目
	BaseRouteType           RouteType `gorm:"column:base_route_typ" json:"base_route_typ,omitempty"`    // 默认轮次
}
