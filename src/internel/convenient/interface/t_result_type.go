package _interface

import (
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
)

type EventID = string
type Player struct {
	WcaID   string `json:"wca_id"`
	WcaName string `json:"WcaName"`

	PlayerId   uint   `json:"PlayerId"`
	CubeId     string `json:"CubeId"`
	PlayerName string `json:"PlayerName"`
}

type PlayerBestResult struct {
	Player
	Single map[EventID]result.Results `json:"Single"`
	Avgs   map[EventID]result.Results `json:"Avgs"`
}

type Nemesis struct {
	PlayerBestResult
}

type KinChSorResultWithEvent struct {
	Event  string
	Result float64
	IsBest bool

	ResultString string // 具体成绩
}

type KinChSorResult struct {
	Player
	Rank    int
	Result  float64
	Results []KinChSorResultWithEvent
}

type UserResultDetail struct {
	RestoresNum  int `json:"RestoresNum"`  // 尝试次数
	SuccessesNum int `json:"SuccessesNum"` // 成功还原次数
	Matches      int `json:"Matches"`      // 比赛场次
	PodiumNum    int `json:"PodiumNum"`    // 领奖台次数
}

type ResultDiff struct {
	A result.Results `json:"A"`
	B result.Results `json:"B"`
}

type AoResults struct {
	EventID EventID          `json:"EventID"`
	Results []result.Results `json:"Results"`
	AoNum   int              `json:"AoNum"`
}

// PlayerEndYears 年终总结
type PlayerEndYears struct {
	Player
	HasYears     []int                  `json:"HasYears"`     // 拥有成绩的年份列表
	CurYear      int                    `json:"CurYear"`      // 当前年份
	CompNum      int                    `json:"CompNum"`      // 比赛次数
	RestoresNum  int                    `json:"RestoresNum"`  // 尝试次数
	SuccessesNum int                    `json:"SuccessesNum"` // 成功还原次数
	Single       map[EventID]ResultDiff `json:"Single"`       // 单次成绩提升
	Avg          map[EventID]ResultDiff `json:"Avg"`          // 平均成绩提升
	Ao12         map[EventID]AoResults  `json:"Ao12"`         // 今年项目最佳ao12
	Ao50         map[EventID]AoResults  `json:"Ao50"`         // 今年项目最佳ao50
	Ao100        map[EventID]AoResults  `json:"Ao100"`        // 今年项目最佳ao100
}
