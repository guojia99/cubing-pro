package result

import (
	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
)

type Record struct {
	basemodel.Model

	EventId     string            `gorm:"column:event_id"`     // 项目ID
	EventRoute  event.RouteType   `gorm:"column:route_type"`   // 项目类型
	ResultId    string            `gorm:"column:result_id"`    // 成绩ID
	UserId      uint              `gorm:"column:user_id"`      // 用户ID
	UserName    string            `gorm:"column:user_name"`    // 用户名
	CompsId     string            `gorm:"column:comps_id"`     // 比赛ID
	CompsName   string            `gorm:"column:comps_name"`   // 比赛名
	CompsGenre  competition.Genre `gorm:"column:comps_genre"`  // 比赛类型
	Best        *float64          `gorm:"column:best"`         // 最佳成绩
	Average     *float64          `gorm:"column:average"`      // 平均成绩
	Repeatedly  *string           `gorm:"column:repeatedly"`   // 多次尝试成绩
	ThisResults string            `gorm:"column:this_results"` // 本次成绩
}
