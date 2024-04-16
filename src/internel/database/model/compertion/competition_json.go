package compertion

import (
	"time"
)

type CompetitionEvent struct {
	EventName     string                // 项目名称
	SingleQualify float64               // 单次资格线
	AvgQualify    float64               // 平均资格线
	HasResults    float64               // 有成绩
	Schedule      []CompetitionSchedule // 赛程
}

type CompetitionSchedule struct {
	Stage          string        // 赛台
	Event          string        // 项目
	StartTime      time.Time     // 开始时间
	EndTime        time.Time     // 结束时间
	Round          string        // 轮次
	Format         string        // 赛制
	Cutoff         time.Duration // 及格线
	TimeLimit      time.Duration // 还原时限
	MinCompetitors int           // 最低限制人数
	Competitors    int           // 人数
}
