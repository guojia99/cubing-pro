package scramble

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type scrambleValue struct {
	Scramble string `json:"scramble"`
	SvgImage string `json:"svgImage"`
}

func (s *scramble) tNoodleCubeScramble(cube string, nums int) ([]string, error) {
	events := s.events()

	if _, ok := events[cube]; !ok {
		return []string{}, fmt.Errorf("cube not found `%s`", cube)
	}
	event := events[cube]
	var cubeKey = cube
	if strings.Contains(cube, "bf") || strings.Contains(cube, "oh") {
		cubeKey = event.PuzzleID
	}

	// 获取colors
	var colors map[string]string
	colorUrl := fmt.Sprintf("%s/frontend/puzzle/%s/colors", s.tNoodleEndpoint, cubeKey)
	if err := utils.HTTPRequestWithJSON(http.MethodGet, colorUrl, nil, nil, nil, &colors); err != nil {
		return []string{}, err
	}

	url := fmt.Sprintf("%s/frontend/puzzle/%s/scramble", s.tNoodleEndpoint, cubeKey)

	var out []string
	for i := 0; i < nums; i++ {
		var val scrambleValue
		err := utils.HTTPRequestWithJSON(http.MethodPost, url, nil, nil, colors, &val)
		if err != nil {
			return []string{}, err
		}
		out = append(out, val.Scramble)
	}

	return out, nil
}

type Event struct {
	ID                    string   `json:"id"`
	Name                  string   `json:"name"`
	PuzzleID              string   `json:"puzzle_id"`
	PuzzleGroupID         string   `json:"puzzle_group_id"`
	FormatIDs             []string `json:"format_ids"`
	CanChangeTimeLimit    bool     `json:"can_change_time_limit"`
	IsTimedEvent          bool     `json:"is_timed_event"`
	IsFewestMoves         bool     `json:"is_fewest_moves"`
	IsMultipleBlindfolded bool     `json:"is_multiple_blindfolded"`
}

func (s *scramble) events() map[string]Event {
	//http://localhost:2014/frontend/data/events
	url := fmt.Sprintf("%s/frontend/data/events", s.tNoodleEndpoint)
	var evs []Event
	if err := utils.HTTPRequestWithJSON(http.MethodGet, url, nil, nil, nil, &evs); err != nil {
		return nil
	}

	var out = make(map[string]Event)
	for _, ev := range evs {
		out[ev.ID] = ev
	}
	return out
}
