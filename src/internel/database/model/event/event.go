package event

import (
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"

	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
)

type RouteType int
type EventName = string

const (
	RouteTypeNot             RouteType = iota // 非比赛项目
	RouteType1rounds                          // "1_r"      // 单轮项目
	RouteType3roundsBest                      // "3_r_b"    // 三轮取最佳
	RouteType3roundsAvg                       // "3_r_a"    // 三轮取平均
	RouteType5roundsBest                      // "5_r_b"    // 五轮取最佳
	RouteType5roundsAvg                       // "5_r_a"    // 五轮取平均
	RouteType5RoundsAvgHT                     // "5_r_a_ht" // 五轮去头尾取平均
	RouteTypeRepeatedly                       // "ry"       // 单轮多次还原项目, 成绩1:还原数; 成绩2:尝试数; 成绩3:时间;
	RouteType3RepeatedlyBest                  // "3ry"      // 三轮尝试多次还原项目 成绩1:还原数; 成绩2:尝试数; 成绩3:时间; 循环3次
)

// WithBest 排名基于最佳成绩
func (r RouteType) WithBest() bool {
	switch r {
	case RouteType1rounds, RouteType3roundsBest, RouteType5roundsBest, RouteTypeRepeatedly, RouteType3RepeatedlyBest:
		return true
	default:
		return false
	}
}

// N 该项目需要的成绩数
func (r RouteType) N() int {
	switch r {
	case RouteType1rounds:
		return 1
	case RouteType3roundsBest, RouteType3roundsAvg, RouteTypeRepeatedly:
		return 3
	case RouteType5roundsBest, RouteType5roundsAvg, RouteType5RoundsAvgHT:
		return 5
	case RouteType3RepeatedlyBest:
		return 9
	default:
		return 0
	}
}

type Event struct {
	basemodel.StringIDModel

	EventI18NJSON string      `gorm:"column:events"`
	EventI18N     []EventI18N `gorm:"-"`
	IsComp        bool        `gorm:"column:is_comp"`     // 比赛项目
	Icon          string      `gorm:"column:icon"`        // Icon
	IconBase64    string      `gorm:"column:icon_base64"` // Icon base64
	IsWCA         bool        `gorm:"column:is_wca"`      // WCA项目
	RouteType     RouteType   `gorm:"column:route_typ"`   // 轮次
}

type EventI18N struct {
	Language string `json:"language"`
	Name     string `json:"name"`
	Long     string `json:"long"`
	Class    string `json:"class"` // 分类
}

func (c *Event) AfterFind(tx *gorm.DB) (err error) {
	return jsoniter.UnmarshalFromString(c.EventI18NJSON, &c.EventI18N)
}
