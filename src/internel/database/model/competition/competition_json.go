package competition

import (
	"errors"
	"strconv"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
)

type CompetitionJson struct {
	Events []CompetitionEvent `json:"Events,omitempty"`
	Cost   CompetitionCost    `json:"Cost,omitempty"`

	TNoodlePath    string `json:"TNoodlePath,omitempty"` // 保存TNoodle打乱内容的地方
	TNoodlePDFPath string `json:"TNoodlePDFPath"`        // pdf
}

type Cost struct {
	Value     float64   `json:"Value,omitempty"`
	StartTime time.Time `json:"StartTime,omitempty"`
	EndTime   time.Time `json:"EndTime,omitempty"`
}

type CompetitionCost struct {
	BaseCost  Cost                       `json:"BaseCost,omitempty"`
	Costs     []Cost                     `json:"Costs,omitempty"`     // 分阶段的报名费
	EventCost map[string]CompetitionCost `json:"EventCost,omitempty"` // 项目的cost
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
	EventName  string          `json:"EventName,omitempty"`  // 项目名称
	EventID    string          `json:"EventID,omitempty"`    // 项目所指ID
	EventRoute event.RouteType `json:"EventRoute,omitempty"` // 项目类型
	IsComp     bool            `json:"IsComp,omitempty"`     // 是否比赛项目

	// 资格线
	SingleQualify     float64 `json:"SingleQualify,omitempty"`     // 单次资格线
	AvgQualify        float64 `json:"AvgQualify,omitempty"`        // 平均资格线
	HasResultsQualify bool    `json:"HasResultsQualify,omitempty"` // 有成绩

	// 赛程
	Schedule []Schedule `json:"Schedule,omitempty"` // 赛程
	Done     bool       `json:"Done,omitempty"`     // 是否已结束
}

func (c *CompetitionEvent) CurRunningSchedule(round interface{}, run *bool) (Schedule, error) {
	var roundNum = -1

	switch data := round.(type) {
	case string:
		roundNum, _ = strconv.Atoi(data)
	}

	for _, schedule := range c.Schedule {
		if run != nil && schedule.IsRunning != *run {
			continue
		}

		if round == schedule.Round || round == schedule.RoundNum || roundNum == schedule.RoundNum || round == "" {
			return schedule, nil
		}
	}
	return Schedule{}, errors.New("没有执行中的轮次")
}

func (c *CompetitionEvent) UpdateSchedule(round interface{}, schedule Schedule) {
	for n := range c.Schedule {
		if c.Schedule[n].Round == round || c.Schedule[n].RoundNum == round {
			c.Schedule[n] = schedule
			break
		}
	}
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

	ActualStartTime time.Time `json:"ActualStartTime,omitempty"` // 实际开始时间
	ActualEndTime   time.Time `json:"ActualEndTime,omitempty"`   // 实际结束时间

	NoRestrictions bool    `json:"NoRestrictions"`         // 无限制
	Cutoff         float64 `json:"Cutoff,omitempty"`       // 及格线
	CutoffNumber   int     `json:"CutoffNumber,omitempty"` // 及格线把数
	TimeLimit      float64 `json:"TimeLimit,omitempty"`    // 还原时限
	//MinCompetitors int     `json:"MinCompetitors,omitempty"` // 最低限制人数

	RoundNum            int    `json:"RoundNum,omitempty"`            // 轮次数字排序
	IsRunning           bool   `json:"IsRunning,omitempty"`           // 是否正在执行的轮次
	FirstRound          bool   `json:"FirstRound,omitempty"`          // 第一轮
	FinalRound          bool   `json:"FinalRound,omitempty"`          // 最后一轮
	AdvancedToThisRound []uint `json:"AdvancedToNextRound,omitempty"` // 本轮晋级的选手

	// 打乱
	NotScramble  bool       `json:"NotScramble,omitempty"`  // 不需要打乱
	Scrambles    [][]string `json:"Scrambles,omitempty"`    // 打乱
	ScrambleNums int        `json:"ScrambleNums,omitempty"` // 打乱数
}
