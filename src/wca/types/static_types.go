package types

import (
	"slices"
	"time"

	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"
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
	// 个人信息
	WcaID   string `gorm:"type:varchar(10)" json:"wcaId"`
	Name    string `gorm:"type:varchar(255)" json:"name"`
	Country string `gorm:"type:varchar(255)" json:"country"`

	// 完成项目的列表, 用逗号隔开
	DoneEventList []string `gorm:"-" json:"doneEventList"`
	DoneEventJSON string   `json:"doneEventJSON"`

	LackNum int `gorm:"type:int" json:"lackNum"` // 缺少项目的数量

	// 完成的开始时间时间
	IsDone    bool       `gorm:"type:bool" json:"isDone"`
	StartTime *time.Time `gorm:"type:datetime" json:"startTime"`
	EndTime   *time.Time `gorm:"type:datetime" json:"endTime"`
	UseDate   int        `gorm:"type:int" json:"useDate"`

	// 完成的那场比赛名
	CompID     string `gorm:"type:varchar(255)" json:"allEventCompId"`
	CompName   string `gorm:"type:varchar(255)" json:"allEventCompName"`
	UseCompNum int    `gorm:"type:int" json:"useCompNum"` // 使用比赛数
}

func (c *AllEventAvgPersonResults) BeforeSave(*gorm.DB) error {
	c.DoneEventJSON, _ = jsoniter.MarshalToString(c.DoneEventList)
	return nil
}
func (c *AllEventAvgPersonResults) AfterFind(*gorm.DB) error {
	_ = jsoniter.UnmarshalFromString(c.DoneEventJSON, &c.DoneEventList)
	return nil
}

// AllEventChampionshipsPodium 大满贯领奖台成绩
type AllEventChampionshipsPodium struct {
	// 个人信息
	WcaID   string `json:"wcaID"`
	WcaName string `json:"wcaName"`
	Country string `json:"country"`

	// 项目
	EventID string `json:"eventID"`
	Best    int    `json:"best"`
	Average int    `json:"average"`

	// 只记录首场成绩
	WorldChampionshipID      string `json:"worldChampionshipID"`
	WorldChampionshipName    string `json:"worldChampionshipName"`
	WorldChampionshipRank    int    `json:"worldChampionshipRank"` // 决赛排名
	WorldChampionshipBest    int    `json:"worldChampionshipBest"`
	WorldChampionshipAverage int    `json:"worldChampionshipAverage"`

	ContinentChampionshipID      string `json:"continentChampionshipID"`
	ContinentChampionshipName    string `json:"continentChampionshipName"`
	ContinentChampionshipRank    int    `json:"continentChampionshipRank"`
	ContinentChampionshipBest    int    `json:"continentChampionshipBest"`
	ContinentChampionshipAverage int    `json:"continentChampionshipAverage"`

	CountryChampionshipID      string `json:"countryChampionshipID"`
	CountryChampionshipName    string `json:"countryChampionshipName"`
	CountryChampionshipRank    int    `json:"countryChampionshipRank"`
	CountryChampionshipBest    int    `json:"countryChampionshipBest"`
	CountryChampionshipAverage int    `json:"countryChampionshipAverage"`

	// WR记录
	HasWR bool `json:"hasWR"`
}

// DiyEventRanks 预计算的 world 排名，每个组合只存前 500
type DiyEventRanks struct {
	EventIndexID uint64 `gorm:"column:event_index_id" json:"eventIndexId"`
	WcaID        string `gorm:"column:wca_id" json:"wcaID"`
	Value        int    `gorm:"column:value" json:"value"` // 排名总和
	Rank         int    `gorm:"column:rank" json:"rank"`   // 排名
	Total        int    `gorm:"column:total" json:"total"` // 总人数
}

type DiyEventSingleRanks struct {
	DiyEventRanks
}

func (DiyEventSingleRanks) TableName() string { return "diy_event_single_ranks" }

type DiyEventAvgRanks struct {
	DiyEventRanks
}

func (DiyEventAvgRanks) TableName() string { return "diy_event_avg_ranks" }

// DiyEventRanksEventIndex 组合与 Events 的映射，id 自增
type DiyEventRanksEventIndex struct {
	ID     uint64
	Events string
}

func (DiyEventRanksEventIndex) TableName() string { return "diy_event_ranks_event_index" }

type RankWithPersonCompStartYear struct {
	PersonID   string `gorm:"column:person_id" json:"personID"`
	PersonName string `gorm:"column:person_name" json:"personName"`
	CountryID  string `gorm:"column:country_id" json:"countryID"`

	Year    int    `gorm:"column:year" json:"year"`
	IsAvg   bool   `gorm:"column:is_avg" json:"isAvg"`
	EventID string `gorm:"column:event_id" json:"eventID"`
	Best    int    `gorm:"column:best" json:"best"`

	WorldRank   int `gorm:"column:world_rank" json:"worldRank"`
	CountryRank int `gorm:"column:country_rank" json:"countryRank"`

	Rank int `gorm:"-" json:"rank"`
}

type PersonPodiums struct {
	PersonID   string `gorm:"column:person_id" json:"personID"`
	PersonName string `gorm:"column:person_name" json:"personName"`
	CountryID  string `gorm:"column:country_id" json:"countryID"`

	BestPodium int16 `gorm:"column:best_podium" json:"bestPodium"`

	HasPodiumEvents     []string `gorm:"-" json:"hasPodiumEvents"`
	HasPodiumEventsJSON string   `gorm:"column:has_podium_events_json" json:"-"`

	// 领奖台
	Gold   int `json:"gold" gorm:"column:gold"`
	Silver int `json:"silver" gorm:"column:silver"`
	Bronze int `json:"bronze" gorm:"column:bronze"`
	Total  int `json:"total" gorm:"column:total"`
}

func (p *PersonPodiums) SetEvent(eventID string) {
	if slices.Contains(p.HasPodiumEvents, eventID) {
		return
	}
	p.HasPodiumEvents = append(p.HasPodiumEvents, eventID)
}

func (p *PersonPodiums) BeforeSave(db *gorm.DB) (err error) {
	p.HasPodiumEventsJSON, err = jsoniter.MarshalToString(p.HasPodiumEvents)
	return
}

func (p *PersonPodiums) AfterFind(db *gorm.DB) (err error) {
	err = jsoniter.UnmarshalFromString(p.HasPodiumEventsJSON, &p.HasPodiumEvents)
	return
}
