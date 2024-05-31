package competition

import (
	"errors"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
)

type CompetitionJson struct {
	Events []CompetitionEvent `json:"Events"`
	Cost   CompetitionCost    `json:"Cost"`
}

type Cost struct {
	Value     float64   `json:"Value"`
	StartTime time.Time `json:"StartTime"`
	EndTime   time.Time `json:"EndTime"`
}

type CompetitionCost struct {
	BaseCost  Cost                       `json:"BaseCost"`
	Costs     []Cost                     `json:"Costs"`     // 分阶段的报名费
	EventCost map[string]CompetitionCost `json:"EventCost"` // 项目的cost
}

func (c CompetitionCost) AllCost(currentTime time.Time, eventKeys []string) float64 {
	totalCost := c.BaseCost.Value

	// Add costs for each phase
	for _, cost := range c.Costs {
		if currentTime.After(cost.StartTime) && currentTime.Before(cost.EndTime) {
			totalCost += cost.Value
		}
	}

	// Add costs for each event
	if len(c.EventCost) == 0 || len(eventKeys) == 0 {
		return totalCost
	}
	for _, key := range eventKeys {
		totalCost += c.EventCost[key].AllCost(currentTime, nil)
	}
	return totalCost
}

type CompetitionEvent struct {
	EventName  string          `json:"EventName,omitempty"` // 项目名称
	EventID    string          `json:"EventID,omitempty"`   // 项目所指ID
	EventRoute event.RouteType `json:"EventRoute"`          // 项目类型
	IsComp     bool            `json:"IsComp,omitempty"`    // 是否比赛项目

	// 资格线
	SingleQualify     float64 `json:"SingleQualify,omitempty"`     // 单次资格线
	AvgQualify        float64 `json:"AvgQualify,omitempty"`        // 平均资格线
	HasResultsQualify bool    `json:"HasResultsQualify,omitempty"` // 有成绩

	// 赛程
	Schedule []Schedule `json:"Schedule,omitempty"` // 赛程
	Done     bool       `json:"Done,omitempty"`     // 是否已结束
}

func (c *CompetitionEvent) CurRunningSchedule(round string) (Schedule, error) {
	for _, schedule := range c.Schedule {
		if !schedule.IsRunning {
			continue
		}
		if round == "" || round == schedule.Round {
			return schedule, nil
		}
	}
	return Schedule{}, errors.New("no running schedule found")
}

type Schedule struct {
	Round string `json:"Round,omitempty"` // 轮次

	Stage       string    `json:"Stage,omitempty"`       // 赛台
	Event       string    `json:"Event,omitempty"`       // 项目
	IsComp      bool      `json:"IsComp,omitempty"`      // 是否比赛项目
	StartTime   time.Time `json:"StartTime,omitempty"`   // 开始时间
	EndTime     time.Time `json:"EndTime,omitempty"`     // 结束时间
	Format      string    `json:"Format,omitempty"`      // 赛制
	Competitors int       `json:"Competitors,omitempty"` // 人数

	ActualStartTime time.Time `json:"ActualStartTime"` // 实际开始时间
	ActualEndTime   time.Time `json:"ActualEndTime"`   // 实际结束时间

	Cutoff         float64 `json:"Cutoff,omitempty"`         // 及格线
	TimeLimit      float64 `json:"TimeLimit,omitempty"`      // 还原时限
	MinCompetitors int     `json:"MinCompetitors,omitempty"` // 最低限制人数

	RoundNum            int    `json:"RoundNum"`            // 轮次数字排序
	IsRunning           bool   `json:"IsRunning"`           // 是否正在执行的轮次
	FirstRound          bool   `json:"FirstRound"`          // 第一轮
	FinalRound          bool   `json:"FinalRound"`          // 最后一轮
	AdvancedToThisRound []uint `json:"AdvancedToNextRound"` // 本轮晋级的选手
}
