package types

import (
	"github.com/vmihailenco/msgpack/v5"
	"gorm.io/gorm"
)

//
//// 项目缩写
//const (
//	E222    = "2"
//	E333    = "3"
//	E333bf  = "3b"
//	E333fm  = "fm"
//	E333ft  = "ft"
//	E333mbf = "mb"
//	E333mbo = "mbo"
//	E333oh  = "oh"
//	E444    = "4"
//	E444bf  = "4b"
//	E555    = "5"
//	E555bf  = "5b"
//	E666    = "6"
//	E777    = "7"
//	EClock  = "c"
//	EMagic  = "mg"
//	EMinx   = "m"
//	EMmagic = "mm"
//	EPyram  = "p"
//	ESkewb  = "s"
//	ESq1    = "q"
//)
//
//var StaticsEventMap = map[string]string{
//	"222":    E222,
//	"333":    E333,
//	"333bf":  E333bf,
//	"333fm":  E333fm,
//	"333ft":  E333ft,
//	"333mbf": E333mbf,
//	"333mbo": E333mbo,
//	"333oh":  E333oh,
//	"444":    E444,
//	"444bf":  E444bf,
//	"555":    E555,
//	"555bf":  E555bf,
//	"666":    E666,
//	"777":    E777,
//	"clock":  EClock,
//	"magic":  EMagic,
//	"minx":   EMinx,
//	"mmagic": EMmagic,
//	"pyram":  EPyram,
//	"skewb":  ESkewb,
//	"sq1":    ESq1,
//}

type StaticPersonRank struct {
	Results       map[string]int `json:"r"` // 成绩
	CountryRank   map[string]int `json:"n"` // 国家排名
	WorldRank     map[string]int `json:"w"` // 世界排名
	ContinentRank map[string]int `json:"c"` // 大洲排名
}

type StaticPersonRankWithTimerRanks struct {
	Single *StaticPersonRank `json:"s"`
	Avg    *StaticPersonRank `json:"a"`
}

// StaticPersonRankWithTimer 基于每个季度去计算的每个玩家的项目排名
type StaticPersonRankWithTimer struct {
	ID    int    `gorm:"primaryKey"`
	WcaID string `gorm:"type:varchar(20)"`
	Year  int    // 年
	Month int    // 月
	//Week      int    // 周

	Country   string `gorm:"type:varchar(64)"` // 国家
	Continent string `gorm:"type:varchar(20)"` // 洲

	RanksBinary []byte                         `json:"-" gorm:"type:BLOB"` // json数据
	Ranks       StaticPersonRankWithTimerRanks `gorm:"-"`
}

func (*StaticPersonRankWithTimer) TableName() string {
	return "static_person_rank_with_timer"
}

func (s *StaticPersonRankWithTimer) BeforeSave(*gorm.DB) error {
	data, err := msgpack.Marshal(s.Ranks)
	if err != nil {
		return err
	}
	s.RanksBinary = data // 改为 []byte 字段
	return nil
}

func (s *StaticPersonRankWithTimer) AfterFind(*gorm.DB) error {
	if s.RanksBinary != nil {
		return msgpack.Unmarshal(s.RanksBinary, &s.Ranks)
	}
	return nil
}
