package scramble

import (
	"strings"

	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
)

const (
	repeatedlyNum int = 40
	backupNum         = 2 // 备打数
)

func (s *scramble) Scramble(event event.Event) ([]string, error) {
	if event.IsWCA {
		if event.BaseRouteType.RouteMap().Repeatedly {
			return s.WCACubeScramble(event.ID, repeatedlyNum)
		}
		return s.WCACubeScramble(event.ID, event.BaseRouteType.RouteMap().Rounds+backupNum)
	}

	switch event.AutoScrambleKey {
	case "FTO":
		return s.AutoScramble(FTOScrambleKey, 25, 30, event.BaseRouteType.RouteMap().Rounds+backupNum), nil
	}

	switch event.ScrambleValue {
	case "333mbf":
		return s.WCACubeScramble("333mbf", repeatedlyNum)
	}

	var evs []string
	if event.ScrambleValue != "" {
		evs = strings.Split(event.ScrambleValue, ",")
	}

	var out []string
	for _, ev := range evs {
		data, err := s.WCACubeScramble(ev, event.BaseRouteType.RouteMap().Rounds+backupNum)
		if err != nil {
			return nil, err
		}
		out = append(out, data...)
	}

	return out, nil
}
