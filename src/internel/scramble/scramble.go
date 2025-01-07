package scramble

import (
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
)

type Scramble interface {
	WCACubeScramble(cube string, nums int) ([]string, error)
	AutoScramble(keys [2][]string, min, max int, group int) []string
	Scramble(event event.Event) ([]string, error)
}

func NewScramble(endpoint string) Scramble {
	return &scramble{
		endpoint: endpoint,
	}
}
