package result

import (
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"

	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

const (
	DNF = -10000 - iota
	DNS
)

type Penalty [][]float64

type Results struct {
	basemodel.Model

	CompetitionID  string    `gorm:"column:comp_id"`         // 比赛ID
	Route          uint      `gorm:"column:route"`           // 轮次
	PersonName     string    `gorm:"column:person_name"`     // 玩家名
	UserID         uint      `gorm:"column:user_id"`         // ID
	Best           float64   `gorm:"column:best"`            // 最佳成绩
	BestRepeatedly float64   `gorm:"column:best_repeatedly"` // 多次尝试的成绩
	Average        float64   `gorm:"column:average"`         // 平均成绩
	ResultJSON     string    `gorm:"column:result_json"`     // 成绩列表JSON
	Result         []float64 `gorm:"-"`                      // 成绩数据
	PenaltyJSON    string    `gorm:"column:penalty_json"`    // 判罚
	Penalty        Penalty   `gorm:"-"`                      // 判罚列表

	EventID    string          `gorm:"column:event_id"`   // 项目
	EventName  string          `gorm:"column:event_name"` // 项目名
	EventRoute event.RouteType `gorm:"column:route_type"` // 项目类型
}

// todo 这里会不会有bug， 单个字段更新时？

func (c *Results) updateSave() error {
	if len(c.Result) != 0 {
		c.ResultJSON, _ = jsoniter.MarshalToString(c.ResultJSON)
	}
	if len(c.Penalty) != 0 {
		c.PenaltyJSON, _ = jsoniter.MarshalToString(c.PenaltyJSON)
	}
	return nil
}

func (c *Results) updateFind() error {
	if len(c.ResultJSON) != 0 {
		_ = jsoniter.UnmarshalFromString(c.ResultJSON, &c.Result)
	}
	if len(c.PenaltyJSON) != 0 {
		_ = jsoniter.UnmarshalFromString(c.PenaltyJSON, &c.Penalty)
	}
	return nil
}

func (c *Results) BeforeCreate(*gorm.DB) error { return c.updateSave() }
func (c *Results) BeforeUpdate(*gorm.DB) error { return c.updateSave() }
func (c *Results) BeforeSave(*gorm.DB) error   { return c.updateSave() }
func (c *Results) AfterFind(*gorm.DB) error    { return c.updateFind() }

func (c *Results) D() bool {
	return utils.TIF[bool](c.EventRoute.RouteMap().WithBest, c.DBest(), c.DAvg())
}
func (c *Results) DBest() bool                  { return c.Best <= DNF }
func (c *Results) DAvg() bool                   { return c.Average <= DNF }
func (c *Results) Update() error                { return c.updateBestAndAvg() }
func (c *Results) IsBest(other Results) bool    { return c.isBest(other) }
func (c *Results) IsBestAvg(other Results) bool { return c.isBestAvg(other) }
