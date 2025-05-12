package event

import (
	"strings"

	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
)

type Event struct {
	basemodel.StringIDModel `table:"-"`
	Idx                     int64     `gorm:"column:idx" json:"idx,omitempty"`                          // 项目排序
	Name                    string    `gorm:"column:name" json:"name,omitempty"`                        // 项目名
	OtherNames              string    `gorm:"column:other_name" json:"otherNames,omitempty"`            // 其他名称
	Cn                      string    `gorm:"column:cn" json:"cn,omitempty"`                            // 中文名
	Class                   string    `gorm:"column:class" json:"class,omitempty"`                      // 分类
	IsComp                  bool      `gorm:"column:is_comp" json:"isComp,omitempty"`                   // 比赛项目
	IsNotCube               bool      `gorm:"column:is_not_cube" json:"isNotCube,omitempty"`            // 非魔方项目
	Icon                    string    `gorm:"column:icon" json:"icon,omitempty" table:"-"`              // Icon
	IconBase64              string    `gorm:"column:icon_base64" json:"iconBase64,omitempty" table:"-"` // Icon base64
	IsWCA                   bool      `gorm:"column:is_wca" json:"isWCA,omitempty"`                     // WCA项目
	BaseRouteType           RouteType `gorm:"column:base_route_typ" json:"base_route_typ,omitempty"`    // 默认轮次

	// 非WCA项目打乱
	ScrambleValue   string `gorm:"column:scramble_value" json:"scrambleValue,omitempty" table:"-"` // 打乱ID []string
	AutoScrambleKey string `gorm:"column:auto_scramble_key" json:"autoScrambleKey,omitempty" table:"-"`
	PuzzleID        string `gorm:"column:puzzle_id" json:"puzzleId,omitempty" table:"-"`
}

func (e *Event) ScrambleValues() []string {
	return strings.Split(e.ScrambleValue, ",")
}
