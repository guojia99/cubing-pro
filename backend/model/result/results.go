package result

import (
	"time"

	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"

	basemodel "github.com/guojia99/cubing-pro/model/base"
)

const MaxResult = float64(time.Hour * 72 / time.Second) // three day

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

	Best    float64 // 最佳成绩
	Average float64 // 平均成绩

	ResultJSON string
	Result     []float64

	PenaltyJSON string
	Penalty     Penalty
}

func (c *Results) update() error {
	var result = make([]float64, len(c.Result))
	copy(result, c.Result)
	// todo
	return nil
}

func (c *Results) AfterFind(tx *gorm.DB) (err error) {
	_ = jsoniter.UnmarshalFromString(c.PenaltyJSON, &c.Penalty)
	err = jsoniter.UnmarshalFromString(c.ResultJSON, &c.Result)
	return err
}
