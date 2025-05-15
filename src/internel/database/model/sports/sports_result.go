package sports

import (
	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
)

// SportResult 体育运动表
type SportResult struct {
	basemodel.Model

	EventID   uint    `gorm:"column:event_id" json:"event_id"`
	EventName string  `gorm:"column:event_name" json:"event_name"`
	UserID    uint    `gorm:"column:user_id" json:"UserID,omitempty"`
	UserName  string  `gorm:"column:user_name" json:"user_name"`
	CubeID    string  `gorm:"column:cube_id" json:"CubeID,omitempty"`
	Result    float64 `gorm:"column:result" json:"Result,omitempty"` // 秒数
	Date      string  `gorm:"column:date" json:"Date,omitempty"`     // 运动日期
	Ban       bool    `gorm:"column:ban" json:"Ban,omitempty"`       // 该成绩是否被ban
	Rank      int     `json:"Rank,omitempty" gorm:"-"`               // 排名
}

// SportEvent 体育运动项目表
type SportEvent struct {
	basemodel.Model `table:"-"`
	Name            string `gorm:"column:name" json:"name"`
	Icon            string `gorm:"column:icon" json:"icon,omitempty" table:"-"`              // Icon
	IconBase64      string `gorm:"column:icon_base64" json:"iconBase64,omitempty" table:"-"` // Icon base64
}
