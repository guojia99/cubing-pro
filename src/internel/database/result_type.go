package database

import (
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
)

type EventID = string
type Player struct {
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
	Event  event.Event
	Result float64
	IsBest bool
}

type KinChSorResult struct {
	Player
	Result  float64
	Results []KinChSorResultWithEvent
}
