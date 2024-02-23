package result

import (
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"

	basemodel "github.com/guojia99/cubing-pro/backend/pkg/model/base"
	"github.com/guojia99/cubing-pro/backend/pkg/utils"
)

const (
	DNF = -10000 - iota
	DNS
)

type Penalty [][]float64

type Results struct {
	basemodel.Model

	CompetitionID string    `gorm:"column:comp_id"`    // 比赛ID
	EvnetID       string    `gorm:"column:event_id"`   // 项目
	EventRoute    RouteType `gorm:"column:route_type"` // 项目类型
	Route         uint      `gorm:"column:route"`      // 轮次

	PersonName string `gorm:"column:person_name"` // 玩家名
	UserID     uint   `gorm:"column:user_id"`     // ID

	Best               float64 `gorm:"column:best"`                 // 最佳成绩
	BestRepeatedlyTime float64 `gorm:"column:best_repeatedly_time"` // 多次尝试的成绩
	Average            float64 `gorm:"column:average"`              // 平均成绩

	ResultJSON string    `gorm:"column:result_json"`
	Result     []float64 `gorm:"-"`

	PenaltyJSON string  `gorm:"column:penalty_json"`
	Penalty     Penalty `gorm:"-"`

	Rank int `gorm:"column:-"`
}

func (c *Results) AfterFind(tx *gorm.DB) (err error) {
	_ = jsoniter.UnmarshalFromString(c.PenaltyJSON, &c.Penalty)
	err = jsoniter.UnmarshalFromString(c.ResultJSON, &c.Result)
	return err
}

func (c *Results) D() bool                      { return utils.TIF[bool](c.EventRoute.WithBest(), c.DBest(), c.DAvg()) }
func (c *Results) DBest() bool                  { return c.Best <= DNF }
func (c *Results) DAvg() bool                   { return c.Average <= DNF }
func (c *Results) Update() error                { return c.updateBestAndAvg() }
func (c *Results) IsBest(other Results) bool    { return c.isBest(other) }
func (c *Results) IsBestAvg(other Results) bool { return c.isBestAvg(other) }
