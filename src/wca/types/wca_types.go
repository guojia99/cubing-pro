package types

import (
	"time"
)

// Championship 代表一个锦标赛记录
type Championship struct {
	ID               int64  `gorm:"column:id;primaryKey;autoIncrement"` // 自增主键 ID
	CompetitionID    string `gorm:"column:competition_id;not null"`     // 关联的比赛 ID
	ChampionshipType string `gorm:"column:championship_type;not null"`  // 锦标赛类型（如 World, Continental 等）
}

func (Championship) TableName() string { return "championships" }

// Competition 代表一场比赛（赛事）
type Competition struct {
	ID                    string `gorm:"column:id;primaryKey;not null" json:"id"`                     // 比赛唯一标识符（如 2025SHAN01）
	Name                  string `gorm:"column:name;not null" json:"name"`                            // 比赛名称
	CityName              string `gorm:"column:city_name;not null" json:"city_name"`                  // 举办城市名
	CountryID             string `gorm:"column:country_id;not null" json:"country_id"`                // 所属国家 ID（关联 countries 表）
	Information           string `gorm:"column:information" json:"information"`                       // 比赛补充信息（mediumtext）
	Year                  uint16 `gorm:"column:year;not null" json:"year"`                            // 开始年份
	Month                 uint16 `gorm:"column:month;not null" json:"month"`                          // 开始月份
	Day                   uint16 `gorm:"column:day;not null" json:"day"`                              // 开始日期
	EndYear               uint16 `gorm:"column:end_year;not null" json:"end_year"`                    // 结束年份
	EndMonth              uint16 `gorm:"column:end_month;not null" json:"end_month"`                  // 结束月份
	EndDay                uint16 `gorm:"column:end_day;not null" json:"end_day"`                      // 结束日期
	Cancelled             int32  `gorm:"column:cancelled;not null" json:"cancelled"`                  // 是否取消（0=未取消，1=已取消）
	EventSpecs            string `gorm:"column:event_specs" json:"event_specs"`                       // 赛事详细配置（JSON 或文本，longtext）
	Delegates             string `gorm:"column:delegates" json:"delegates"`                           // WCA 代表列表（mediumtext）
	Organizers            string `gorm:"column:organizers" json:"organizers"`                         // 组织者列表（mediumtext）
	Venue                 string `gorm:"column:venue;not null" json:"venue"`                          // 场馆名称
	VenueAddress          string `gorm:"column:venue_address" json:"venue_address"`                   // 场馆地址
	VenueDetails          string `gorm:"column:venue_details" json:"venue_details"`                   // 场馆内部详情（如房间号）
	ExternalWebsite       string `gorm:"column:external_website" json:"external_website"`             // 外部官网链接
	CellName              string `gorm:"column:cell_name;not null" json:"cell_name"`                  // 表格中显示的单元格名称（用于导出等）
	LatitudeMicrodegrees  int32  `gorm:"column:latitude_microdegrees" json:"latitude_microdegrees"`   // 纬度（微度，1度 = 1,000,000 微度）
	LongitudeMicrodegrees int32  `gorm:"column:longitude_microdegrees" json:"longitude_microdegrees"` // 经度（微度，1度 = 1,000,000 微度）

	//拓展参数
	EventIds    []string `json:"event_ids" gorm:"-"`
	CountryIso2 string   `json:"country_iso_2" gorm:"-"`
}

func (Competition) TableName() string { return "competitions" }

// Continent 代表大洲信息
type Continent struct {
	ID         string `gorm:"column:id;primaryKey;not null"` // 大洲唯一标识（如 asia, europe）
	Name       string `gorm:"column:name;not null"`          // 大洲中文/英文名称
	RecordName string `gorm:"column:record_name;not null"`   // 记录缩写（如 As, Eu, Af，用于成绩记录）
}

func (Continent) TableName() string { return "continents" }

// Country 代表国家信息
type Country struct {
	ID          string `gorm:"column:id;primaryKey;not null"` // 国家唯一标识（如 China, USA）
	Name        string `gorm:"column:name;not null"`          // 国家名称
	ContinentID string `gorm:"column:continent_id;not null"`  // 所属大洲 ID
	ISO2        string `gorm:"column:iso2"`                   // ISO 3166-1 alpha-2 两位国家代码（如 CN, US）
}

func (Country) TableName() string { return "countries" }

// EligibleCountryISO2ForChampionship 某类锦标赛允许参赛的国家 ISO2 列表
type EligibleCountryISO2ForChampionship struct {
	ChampionshipType    string `gorm:"column:championship_type;not null"`     // 锦标赛类型
	EligibleCountryISO2 string `gorm:"column:eligible_country_iso2;not null"` // 允许参赛的国家 ISO2 代码
}

func (EligibleCountryISO2ForChampionship) TableName() string {
	return "eligible_country_iso2s_for_championship"
}

// Event 代表魔方项目（如 333, 444, oh 等）
type Event struct {
	ID     string `gorm:"column:id;primaryKey;not null"` // 项目 ID（如 333, 444, minx）
	Name   string `gorm:"column:name;not null"`          // 项目全称（如 "3x3x3 Cube"）
	Rank   int32  `gorm:"column:rank;not null"`          // 排序优先级（数值越小越靠前）
	Format string `gorm:"column:format;not null"`        // 默认成绩格式（如 a, m, s）
}

func (Event) TableName() string { return "events" }

// Format 代表成绩格式定义
type Format struct {
	ID                 string `gorm:"column:id;primaryKey;not null"`        // 格式 ID（如 a=平均, m=多次取最佳, s=单次）
	Name               string `gorm:"column:name;not null"`                 // 格式名称（如 "Average of 5"）
	SortBy             string `gorm:"column:sort_by;not null"`              // 主排序字段（如 "average", "best"）
	SortBySecond       string `gorm:"column:sort_by_second;not null"`       // 次排序字段（用于平局）
	ExpectedSolveCount int32  `gorm:"column:expected_solve_count;not null"` // 预期尝试次数（如 5 次）
	TrimFastestN       int32  `gorm:"column:trim_fastest_n;not null"`       // 去除最快 N 次（通常为 0 或 1）
	TrimSlowestN       int32  `gorm:"column:trim_slowest_n;not null"`       // 去除最慢 N 次（通常为 0 或 1）
}

func (Format) TableName() string { return "formats" }

// Person 代表选手信息
type Person struct {
	WcaID     string `gorm:"column:wca_id;not null"`     // WCA 选手 ID（如 2020XXXX01）
	SubID     int8   `gorm:"column:sub_id;not null"`     // 子账号 ID（用于区分同名选手，通常为 1）
	Name      string `gorm:"column:name"`                // 选手姓名
	CountryID string `gorm:"column:country_id;not null"` // 所属国家 ID
	Gender    string `gorm:"column:gender"`              // 性别（M/F/U）
}

func (Person) TableName() string { return "persons" }

// RanksAverage 选手在各项目的平均成绩世界/洲/国排名
type RanksAverage struct {
	PersonID      string `gorm:"column:person_id;not null"`      // 选手 WCA ID
	EventID       string `gorm:"column:event_id;not null"`       // 项目 ID
	Best          int    `gorm:"column:best;not null"`           // 最佳平均成绩（单位：百分之一秒）
	WorldRank     int    `gorm:"column:world_rank;not null"`     // 世界排名
	ContinentRank int    `gorm:"column:continent_rank;not null"` // 大洲排名
	CountryRank   int    `gorm:"column:country_rank;not null"`   // 国家排名
}

func (RanksAverage) TableName() string { return "ranks_average" }

// RanksSingle 选手在各项目的单次成绩排名
type RanksSingle struct {
	PersonID      string `gorm:"column:person_id;not null"`      // 选手 WCA ID
	EventID       string `gorm:"column:event_id;not null"`       // 项目 ID
	Best          int    `gorm:"column:best;not null"`           // 最佳单次成绩（单位：百分之一秒）
	WorldRank     int    `gorm:"column:world_rank;not null"`     // 世界排名
	ContinentRank int    `gorm:"column:continent_rank;not null"` // 大洲排名
	CountryRank   int    `gorm:"column:country_rank;not null"`   // 国家排名
}

func (RanksSingle) TableName() string { return "ranks_single" }

// ResultAttempt 单次尝试的成绩明细
type ResultAttempt struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"` // 自增 ID
	Value         int64     `gorm:"column:value;not null"`              // 成绩值（单位：百分之一秒；DNF=-1, DNS=-2）
	AttemptNumber int16     `gorm:"column:attempt_number;not null"`     // 第几次尝试（1~5）
	ResultID      int64     `gorm:"column:result_id;not null"`          // 所属结果 ID（关联 results 表）
	CreatedAt     time.Time `gorm:"column:created_at;not null"`         // 创建时间（微秒精度）
	UpdatedAt     time.Time `gorm:"column:updated_at;not null"`         // 更新时间（微秒精度）
}

func (ResultAttempt) TableName() string { return "result_attempts" }

// Result 代表一次比赛中的某位选手在某个项目某轮的成绩汇总
type Result struct {
	ID                    int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                  // 自增 ID
	CompetitionID         string `gorm:"column:competition_id;not null" json:"competition_id"`          // 比赛 ID
	EventID               string `gorm:"column:event_id;not null" json:"event_id"`                      // 项目 ID
	RoundTypeID           string `gorm:"column:round_type_id;not null" json:"round_type_id"`            // 轮次类型（如 f=决赛, s=半决赛）
	Pos                   int16  `gorm:"column:pos;not null" json:"pos"`                                // 排名位置（1 为第一名）
	Best                  int    `gorm:"column:best;not null" json:"best"`                              // 最佳单次成绩
	Average               int    `gorm:"column:average;not null" json:"average"`                        // 平均成绩
	PersonName            string `gorm:"column:person_name" json:"name"`                                // 选手姓名（冗余字段）
	PersonID              string `gorm:"column:person_id;not null" json:"wca_id"`                       // 选手 WCA ID
	PersonCountryID       string `gorm:"column:person_country_id" json:"country_iso2"`                  // 选手所属国家（冗余）
	FormatID              string `gorm:"column:format_id;not null" json:"format_id"`                    // 成绩格式 ID
	RegionalSingleRecord  string `gorm:"column:regional_single_record" json:"regional_single_record"`   // 区域单次纪录（如 WR, CR, NR）
	RegionalAverageRecord string `gorm:"column:regional_average_record" json:"regional_average_record"` // 区域平均纪录（如 WR, CR, NR）

	// 拓展参数
	Attempts   []int64 `gorm:"-" json:"attempts"`
	BestIndex  int     `gorm:"-" json:"best_index"` // 最佳成绩在 attempts 中的索引（从 0 开始）
	WorstIndex int     `gorm:"-" json:"worst_index"`
}

func (Result) TableName() string { return "results" }

// RoundType 轮次类型定义（如初赛、决赛等）
type RoundType struct {
	ID       string `gorm:"column:id;primaryKey;not null"` // 轮次类型 ID（如 c, s, f）
	Rank     int32  `gorm:"column:rank;not null"`          // 排序优先级（数值越小越早进行）
	Name     string `gorm:"column:name;not null"`          // 轮次名称（如 "Final", "Semi Final"）
	CellName string `gorm:"column:cell_name;not null"`     // 表格中显示的简写（如 "f", "sf"）
	Final    bool   `gorm:"column:final;not null"`         // 是否为决赛轮
}

func (RoundType) TableName() string { return "round_types" }

// SchemaMigration 数据库迁移版本记录
type SchemaMigration struct {
	Version string `gorm:"column:version;primaryKey;not null"` // 迁移版本号（如 "20230101000000"）
}

func (SchemaMigration) TableName() string { return "schema_migrations" }

// Scramble 比赛中的打乱序列
type Scramble struct {
	ID            int64  `gorm:"column:id;primaryKey;autoIncrement"` // 自增 ID
	CompetitionID string `gorm:"column:competition_id;not null"`     // 比赛 ID
	EventID       string `gorm:"column:event_id;not null"`           // 项目 ID
	RoundTypeID   string `gorm:"column:round_type_id;not null"`      // 轮次类型
	GroupID       string `gorm:"column:group_id;not null"`           // 分组 ID（如 "1", "2", "A"）
	IsExtra       bool   `gorm:"column:is_extra;not null"`           // 是否为加试打乱（extra attempt）
	ScrambleNum   int32  `gorm:"column:scramble_num;not null"`       // 打乱序号（1~5）
	Scramble      string `gorm:"column:scramble;not null"`           // 打乱公式（字符串）
}

func (Scramble) TableName() string { return "scrambles" }
