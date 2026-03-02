package types

import (
	"time"
)

type StaticWithTimerRank struct {
	WcaID   string `gorm:"type:varchar(10)" json:"wcaId"`
	WcaName string `gorm:"-" json:"wcaName"`
	EventID string `gorm:"type:varchar(10)" json:"eventId"`
	Year    int    `gorm:"type:smallint unsigned" json:"year"`
	Month   int    `gorm:"type:tinyint unsigned" json:"month"`
	Week    int    `gorm:"type:tinyint unsigned" json:"week"`
	Single  int    `gorm:"type:int" json:"single"`
	Average int    `gorm:"type:int" json:"average"`

	Country string `gorm:"type:varchar(255)" json:"country"`

	SingleCountryRank   int `gorm:"type:mediumint unsigned" json:"singleCountryRank"`
	SingleWorldRank     int `gorm:"type:mediumint unsigned" json:"singleWorldRank"`
	SingleContinentRank int `gorm:"type:mediumint unsigned" json:"singleContinentRank"`

	AvgCountryRank   int `gorm:"type:mediumint unsigned" json:"avgCountryRank"`
	AvgWorldRank     int `gorm:"type:mediumint unsigned" json:"avgWorldRank"`
	AvgContinentRank int `gorm:"type:mediumint unsigned" json:"avgContinentRank"`
}

type StaticSuccessRateResult struct {
	WcaID   string `gorm:"type:varchar(10)" json:"wcaId"`
	WcaName string `gorm:"type:varchar(255)" json:"wcaName"`
	Country string `gorm:"type:varchar(255)" json:"country"`

	EventID string `gorm:"type:varchar(10)" json:"eventId"`

	Solved     int     `gorm:"type:int" json:"solved"`       // 数量
	Attempted  int     `gorm:"type:int" json:"attempted"`    // 尝试
	Percentage float64 `gorm:"type:float" json:"percentage"` // 成功率
}

// AllEventAvgPersonResults 全项目有平均 -- 粗饼"大满贯"
type AllEventAvgPersonResults struct {
	WcaID   string `gorm:"type:varchar(10)" json:"wcaId"`
	Name    string `gorm:"type:varchar(255)" json:"name"`
	Country string `gorm:"type:varchar(255)" json:"country"`

	// 完成项目的列表, 用逗号隔开
	DoneEventList string `gorm:"type:varchar(255)" json:"doneEventList"`
	LackNum       int    `gorm:"type:int" json:"lackNum"` // 缺少某些项目的数量

	// 完成的开始时间时间
	IsDone    bool      `gorm:"type:bool" json:"isDone"`
	StartTime time.Time `gorm:"type:datetime" json:"startTime"`
	EndTime   time.Time `gorm:"type:datetime" json:"endTime"`

	// 完成的那场比赛名
	CompID   string `gorm:"type:varchar(255)" json:"allEventCompId"`
	CompName string `gorm:"type:varchar(255)" json:"allEventCompName"`
}
