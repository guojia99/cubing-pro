package competition

import (
	"time"
)

type CompetitionEvent struct {
	EventName     string                `json:"EventName,omitempty"`     // 项目名称
	SingleQualify float64               `json:"SingleQualify,omitempty"` // 单次资格线
	AvgQualify    float64               `json:"AvgQualify,omitempty"`    // 平均资格线
	HasResults    float64               `json:"HasResults,omitempty"`    // 有成绩
	Schedule      []CompetitionSchedule `json:"Schedule,omitempty"`      // 赛程
}

type CompetitionSchedule struct {
	Stage          string        `json:"Stage,omitempty"`          // 赛台
	Event          string        `json:"Event,omitempty"`          // 项目
	EventID        string        `json:"EventID,omitempty"`        // 项目所指ID
	IsComp         bool          `json:"IsComp,omitempty"`         // 是否比赛项目
	StartTime      time.Time     `json:"StartTime,omitempty"`      // 开始时间
	EndTime        time.Time     `json:"EndTime,omitempty"`        // 结束时间
	Round          string        `json:"Round,omitempty"`          // 轮次
	Format         string        `json:"Format,omitempty"`         // 赛制
	Cutoff         time.Duration `json:"Cutoff,omitempty"`         // 及格线
	TimeLimit      time.Duration `json:"TimeLimit,omitempty"`      // 还原时限
	MinCompetitors int           `json:"MinCompetitors,omitempty"` // 最低限制人数
	Competitors    int           `json:"Competitors,omitempty"`    // 人数
}
