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

	CompetitionID string    // 比赛ID
	EvnetID       string    // 项目
	EventRoute    RouteType // 项目类型
	Route         uint      // 轮次

	PersonName string // 玩家名
	PersonId   string // ID

	Best               float64 // 最佳成绩
	BestRepeatedlyTime float64 // 多次尝试的成绩
	Average            float64 // 平均成绩

	ResultJSON string
	Result     []float64

	PenaltyJSON string
	Penalty     Penalty

	Rank int
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

/*
W>Y>Q>S 起手是R

SY	R U' (R' E R2 E' R') U R'
YS	R U' (R E R2 E' R) U R'
SQ	R' U' (R' E R2 E' R') U R
QS	R' U' (R E R2 E' R) U R
SW	U' (r E r2' E r) U
WS	U' (r' E' r2 E' r') U
QW	R' U' (R' E' R2 E R') U R
WQ	R' U' (R E' R2 E R) U R
QY	U (R' S R2 S' R') U'
YQ	U (R S R2 S' R) U'
YW	R U' (R' E' R2 E R') U R'
WY	R U' (R E' R2 E R) U R'
TR	U E (R' S R2 S' R') U' E'
RT	U E (R S R2 S' R) U' E'
TZ	E R U' R' E2 R U R' E
ZT	E' R U' R' E2 R U R' E'
RZ	E R U' (R' E' R2 E R') U R' E'
ZR	E R U' (R E' R2 E R) U R' E'
RX	r U r' E2 r U' r' E2
XR	E2 r U r' E2 r U' r'
XZ	F R2 E' R2 E F'
ZX	F E' R2 E R2 F'
*/
