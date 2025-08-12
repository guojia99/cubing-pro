package pktimer

import (
	"time"

	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/robot/types"
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"
)

type Player struct {
	QQ       int64  `json:"QQ"`
	QQBot    string `json:"QQBot"`
	UserName string `json:"userName"`
	UserId   uint   `json:"userId"`

	Results []float64 `json:"results"`

	Best    float64 `json:"best"`
	Average float64 `json:"average"`

	// 退出
	Exit    bool `json:"exit"`
	ExitNum int  `json:"exitNum"`
}

type PkResults struct {
	Players      []Player        `json:"players"`
	Event        event.Event     `json:"event"`
	Count        int             `json:"count"`        // 轮次
	CurCount     int             `json:"curCount"`     // 当前轮次
	FirstMessage types.InMessage `json:"firstMessage"` // 上次消息
}

type PkTimerResult struct {
	basemodel.Model

	GroupID     string    `json:"groupId"`
	Running     bool      `json:"running"`
	Start       bool      // 开始比赛
	LastRunning time.Time `json:"lastRunning"` // 上一次更新时间
	StartPerson string    `json:"startPerson"` // 开启的

	ResultsJSON string
	PkResults   PkResults `json:"pkResults" gorm:"-"`
	Eps         float64   `json:"eps"` // 发货精度

	// 成功结束
	//SuccessDone bool // 正常结束的
}

func (c *PkTimerResult) BeforeSave(*gorm.DB) error {
	c.ResultsJSON, _ = jsoniter.MarshalToString(c.PkResults)
	return nil
}
func (c *PkTimerResult) AfterFind(*gorm.DB) error {
	_ = jsoniter.UnmarshalFromString(c.ResultsJSON, &c.PkResults)
	return nil
}
