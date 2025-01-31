package scramble

import (
	"strings"

	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
)

type Scramble interface {
	ScrambleWithComp(event event.Event) ([]string, error)
	Scramble(ev string, num int) []string
	Test() string
}

func NewScramble(scrambleType string, tNoodleEndpoint string) Scramble {
	if scrambleType == "" {
		scrambleType = scrambleTypeTNoodle
	}
	s := &scramble{
		scrambleType:    scrambleType,
		tNoodleEndpoint: tNoodleEndpoint,
	}
	if s.scrambleType == scrambleTypeRustTwisty {
		go s.loopRustScrambleCache()
	}

	return s
}

type scramble struct {
	scrambleType    string
	tNoodleEndpoint string
}

func (s *scramble) Test() string {
	switch s.scrambleType {
	case scrambleTypeTNoodle:
		return ""
	case scrambleTypeRustTwisty:
		return s.rustTestLongScramble()
	}
	return ""
}

const (
	repeatedlyNum int = 99
	backupNum         = 2 // 备打数
)

const (
	scrambleTypeRustTwisty = "rust_twisty" // 狼的打乱
	scrambleTypeTNoodle    = "tnoodle"
)

func (s *scramble) ScrambleWithComp(event event.Event) ([]string, error) {
	if event.IsWCA {
		if event.BaseRouteType.RouteMap().Repeatedly {
			return s.Scramble(event.ID, repeatedlyNum), nil
		}
		return s.Scramble(event.ID, event.BaseRouteType.RouteMap().Rounds+backupNum), nil
	}

	switch event.AutoScrambleKey {
	case "FTO":
		return s.autoScramble(FTOScrambleKey, 25, 30, event.BaseRouteType.RouteMap().Rounds+backupNum), nil
	}

	switch event.ScrambleValue {
	case "333mbf":
		return s.Scramble("333mbf", repeatedlyNum), nil
	}

	var evs []string
	if event.ScrambleValue != "" {
		evs = strings.Split(event.ScrambleValue, ",")
	}

	var out []string
	for _, ev := range evs {
		data := s.Scramble(ev, event.BaseRouteType.RouteMap().Rounds+backupNum)
		out = append(out, data...)
	}

	return out, nil
}

func (s *scramble) Scramble(ev string, num int) []string {
	var wcaHandler func(string, int) ([]string, error)
	switch s.scrambleType {
	case scrambleTypeTNoodle:
		wcaHandler = s.tNoodleCubeScramble
	case scrambleTypeRustTwisty:
		wcaHandler = s.rustScramble
	}
	if wcaHandler == nil {
		//, fmt.Errorf("scramble type %s not supported", s.scrambleType)
		return nil
	}
	data, err := wcaHandler(ev, num)
	if err != nil {
		return nil
	}

	return data
}
