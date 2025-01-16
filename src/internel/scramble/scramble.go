package scramble

import (
	"fmt"
	"strings"

	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
)

type Scramble interface {
	TNoodleCubeScramble(cube string, nums int) ([]string, error)
	AutoScramble(keys [2][]string, min, max int, group int) []string
	Scramble(event event.Event) ([]string, error)
}

func NewScramble(scrambleType string, tNoodleEndpoint string) Scramble {
	if scrambleType == "" {
		scrambleType = scrambleTypeTNoodle
	}

	return &scramble{
		scrambleType:    scrambleType,
		tNoodleEndpoint: tNoodleEndpoint,
	}
}

type scramble struct {
	scrambleType string

	tNoodleEndpoint string
}

const (
	repeatedlyNum int = 40
	backupNum         = 2 // 备打数
)

const (
	scrambleTypeRustTwisty = "rust_twisty" // 狼的打乱
	scrambleTypeTNoodle    = "tnoodle"
)

func (s *scramble) Scramble(event event.Event) ([]string, error) {
	var handler func(string, int) ([]string, error)
	switch s.scrambleType {
	case scrambleTypeTNoodle:
		handler = s.TNoodleCubeScramble
	case scrambleTypeRustTwisty:
		handler = s.LangScramble
	}
	if handler == nil {
		return nil, fmt.Errorf("scramble type %s not supported", s.scrambleType)
	}

	if event.IsWCA {
		if event.BaseRouteType.RouteMap().Repeatedly {
			return handler(event.ID, repeatedlyNum)
		}
		return handler(event.ID, event.BaseRouteType.RouteMap().Rounds+backupNum)
	}

	switch event.AutoScrambleKey {
	case "FTO":
		return s.AutoScramble(FTOScrambleKey, 25, 30, event.BaseRouteType.RouteMap().Rounds+backupNum), nil
	}

	switch event.ScrambleValue {
	case "333mbf":
		return handler("333mbf", repeatedlyNum)
	}

	var evs []string
	if event.ScrambleValue != "" {
		evs = strings.Split(event.ScrambleValue, ",")
	}

	var out []string
	for _, ev := range evs {
		data, err := handler(ev, event.BaseRouteType.RouteMap().Rounds+backupNum)
		if err != nil {
			return nil, err
		}
		out = append(out, data...)
	}

	return out, nil
}
