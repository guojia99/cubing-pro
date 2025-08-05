package scramble

import (
	"fmt"
	"strings"

	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
)

type Scramble interface {
	ScrambleWithComp(event event.Event) ([]string, error)
	ScrambleWithEvent(event event.Event, number int) ([]string, error)
	Scramble(ev string, num int) []string
	Test() string
	Image(scramble string, ev string) (string, error)
}

func NewScramble(scrambleType string, tNoodleEndpoint string, scrambleDrawType string, scrambleUrl string) Scramble {
	if scrambleType == "" {
		scrambleType = scrambleTypeTNoodle
	}
	if scrambleDrawType == "" {
		scrambleDrawType = scrambleTypeDrawType2Mf8
	}

	s := &scramble{
		scrambleType:     scrambleType,
		tNoodleEndpoint:  tNoodleEndpoint,
		scrambleDrawType: scrambleDrawType,
		scrambleUrl:      scrambleUrl,
	}
	if s.scrambleType == scrambleTypeRustTwisty {
		//go s.loopRustScrambleCache()
	}

	return s
}

type scramble struct {
	scrambleType    string
	tNoodleEndpoint string

	scrambleDrawType string // 2mf8
	scrambleUrl      string
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
	scrambleTypeRustTwisty   = "rust_twisty" // 狼的打乱
	scrambleTypeTNoodle      = "tnoodle"
	scrambleTypeDrawType2Mf8 = "2mf8"
)

func (s *scramble) ScrambleWithEvent(event event.Event, number int) ([]string, error) {
	if event.IsWCA {
		return s.Scramble(event.ID, number), nil
	}

	switch event.AutoScrambleKey {
	case "FTO":
		return s.autoScramble(FTOScrambleKey, 25, 30, number), nil
	}

	var evs []string
	if event.ScrambleValue != "" {
		evs = strings.Split(event.ScrambleValue, ",")
	}

	var out []string
	for i := 0; i < event.BaseRouteType.RouteMap().Rounds; i++ {
		for _, ev := range evs {
			data := s.Scramble(ev, 1)
			out = append(out, data...)
		}
	}
	return out, nil
}

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
		return s.Scramble("333bf", repeatedlyNum), nil
	case "444", "444bf":
		// 等狼优化到极致速度再采用随机状态
		return s.autoScramble(Cube444ScrambleKey, 39, 49, event.BaseRouteType.RouteMap().Rounds+backupNum), nil
	}

	var evs []string
	if event.ScrambleValue != "" {
		evs = strings.Split(event.ScrambleValue, ",")
	}

	var out []string
	for i := 0; i < event.BaseRouteType.RouteMap().Rounds; i++ {
		for _, ev := range evs {
			data := s.Scramble(ev, 1)
			out = append(out, data...)
		}
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

func (s *scramble) Image(scramble string, ev string) (string, error) {
	switch s.scrambleDrawType {
	case scrambleTypeDrawType2Mf8:
		return s.SImageWith2mf8(scramble, ev)
	default:
		return "", fmt.Errorf("scramble draw type %s not supported", s.scrambleDrawType)
	}
}
