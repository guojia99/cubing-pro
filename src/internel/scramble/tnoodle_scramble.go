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

type (
	WcIfExtensionsData struct {
		IsStaging          bool `json:"isStaging"`
		IsManual           bool `json:"isManual"`
		IsSignedBuild      bool `json:"isSignedBuild"`
		IsAllowedVersion   bool `json:"isAllowedVersion"`
		NumCopies          int  `json:"numCopies,omitempty"`
		RequestedScrambles int  `json:"requestedScrambles,omitempty"`
	}

	WcIfExtension struct {
		Id      string             `json:"id"`
		SpecUrl string             `json:"specUrl"`
		Data    WcIfExtensionsData `json:"data"`
	}

	WcIfSchedule struct {
		NumberOfDays int           `json:"numberOfDays"`
		Venues       []interface{} `json:"venues"`
	}

	WcIfRound struct {
		Format           string `json:"format"`
		Id               string `json:"id"`
		ScrambleSetCount int    `json:"scrambleSetCount"`
		Extensions       []struct {
			Id      string `json:"id"`
			SpecUrl string `json:"specUrl"`
			Data    struct {
				NumCopies          int `json:"numCopies,omitempty"`
				RequestedScrambles int `json:"requestedScrambles,omitempty"`
			} `json:"data"`
		} `json:"extensions"`
	}
	WcIfEvent struct {
		Id         string          `json:"id"`
		Rounds     []WcIfRound     `json:"rounds"`
		Extensions []WcIfExtension `json:"extensions"`
	}

	WcIf struct {
		FormatVersion string          `json:"formatVersion"`
		Name          string          `json:"name"`
		ShortName     string          `json:"shortName"`
		Id            string          `json:"id"`
		Events        []WcIfEvent     `json:"events"`
		Schedule      WcIfSchedule    `json:"schedule"`
		Extensions    []WcIfExtension `json:"extensions"`
	}

	tNoodleGenerateScramblesRequest struct {
		WcIf WcIf `json:"wcif"`
	}

	tNoodleGenerateScramblesRequestResp struct {
		// 原始数据
		ZipData []byte `json:"zipData"`
		ZipPath string `json:"zipPath"`

		//
	}
)

func (s *scramble) GenerateScrambles() {}
