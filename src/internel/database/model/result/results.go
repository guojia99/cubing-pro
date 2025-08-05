package result

import (
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"

	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

const (
	DNF = -10000 - iota // 未还原
	DNS                 // 未开始
	DNP                 // 未晋级
	DNT                 // 超时
	UNT                 // 未知
)

type Penalty [][]float64

type Results struct {
	basemodel.Model

	CompetitionID   uint    `gorm:"column:comp_id" json:"CompetitionID,omitempty"`     // 比赛ID
	CompetitionName string  `gorm:"column:comp_name" json:"CompetitionName,omitempty"` // 比赛名
	Round           string  `gorm:"column:round" json:"Round,omitempty"`               // 轮次名
	RoundNumber     int     `gorm:"column:round_number" json:"RoundNumber,omitempty"`  // 轮次号
	PersonName      string  `gorm:"column:person_name" json:"PersonName,omitempty"`    // 玩家名
	UserID          uint    `gorm:"column:user_id" json:"UserID,omitempty"`            // ID
	CubeID          string  `gorm:"column:cube_id" json:"CubeID,omitempty"`            // CubeID
	Best            float64 `gorm:"column:best" json:"Best,omitempty"`                 // 最佳成绩
	Average         float64 `gorm:"column:average" json:"Average,omitempty"`           // 平均成绩

	// 计次项目
	BestRepeatedlyReduction float64 `gorm:"column:best_repeatedly_reduction" json:"BestRepeatedlyReduction,omitempty"` // 计次最佳成绩成功
	BestRepeatedlyTry       float64 `gorm:"column:best_repeatedly_try" json:"BestRepeatedlyTry,omitempty"`             // 计次尝试
	BestRepeatedlyTime      float64 `gorm:"column:best_repeatedly" json:"BestRepeatedlyTime,omitempty"`                // 计次的成绩

	ResultJSON  string    `gorm:"column:result_json" json:"ResultJSON"`             // 成绩列表JSON
	Result      []float64 `gorm:"-" json:"Result,omitempty"`                        // 成绩数据
	PenaltyJSON string    `gorm:"column:penalty_json" json:"PenaltyJSON,omitempty"` // 判罚
	Penalty     Penalty   `gorm:"-" json:"Penalty,omitempty"`                       // 判罚列表

	EventID    string          `gorm:"column:event_id" json:"EventID,omitempty"`      // 项目
	EventName  string          `gorm:"column:event_name" json:"EventName,omitempty"`  // 项目名
	EventRoute event.RouteType `gorm:"column:route_type" json:"EventRoute,omitempty"` // 项目类型
	Ban        bool            `gorm:"column:ban" json:"Ban,omitempty"`               // 该成绩是否被ban

	Rank int `json:"Rank,omitempty" gorm:"-"` // 排名
}

// todo 这里会不会有bug， 单个字段更新时？

func (c *Results) updateSave() error {
	if len(c.Result) != 0 {
		c.Result = c.Result[:c.EventRoute.RouteMap().Rounds]
		c.ResultJSON, _ = jsoniter.MarshalToString(c.Result)
	}
	if len(c.Penalty) != 0 {
		c.PenaltyJSON, _ = jsoniter.MarshalToString(c.Penalty)
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
func (c *Results) BestString() string           { return c.bestString() }
func (c *Results) BestAvgString() string        { return c.bestAvgString() }
func (c *Results) EqualRepeatedly(other Results) bool {
	return c.BestRepeatedlyReduction == other.BestRepeatedlyReduction &&
		c.BestRepeatedlyTry == other.BestRepeatedlyTry &&
		c.BestRepeatedlyTime == other.BestRepeatedlyTime
}
