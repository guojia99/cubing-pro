package groupcomp

import (
	"errors"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"math"
	"regexp"
	"strconv"
	"strings"
)

func MatchesGroup(field, id string) bool {
	if id == "" || id == "<nil>" {
		return false
	}
	for _, v := range competition.StringListFromDB(field) {
		if strings.TrimSpace(v) == id {
			return true
		}
	}
	return false
}

var batchSeparator = regexp.MustCompile(`\s+/\s+|[\n\r]+`)
var resultToken = regexp.MustCompile(`^\d+(?::\d+){0,2}(?:\.\d+)?$`)

func SplitBatch(msg string) []string { return batchSeparator.Split(strings.TrimSpace(msg), -1) }

func ParseValues(tokens []string, route event.RouteType) ([]float64, error) {
	m := route.RouteMap()
	if len(tokens) == 0 || m.Rounds < 1 {
		return nil, errors.New("成绩不能为空或赛制无效")
	}
	var values []float64
	for i := 0; i < len(tokens); i++ {
		if m.Repeatedly {
			parts := strings.Split(tokens[i], "/")
			if len(parts) != 2 || i+1 >= len(tokens) {
				return nil, errors.New("多盲格式须为成功数/尝试数 时间")
			}
			a, err := strconv.Atoi(parts[0])
			if err != nil {
				return nil, err
			}
			b, err := strconv.Atoi(parts[1])
			if err != nil || a < 0 || b < 1 || a > b {
				return nil, errors.New("多盲成功数/尝试数无效")
			}
			values = append(values, float64(a), float64(b))
			i++
		}
		token := strings.ToLower(tokens[i])
		if token != "d" && token != "dnf" && token != "s" && token != "dns" && !resultToken.MatchString(token) {
			return nil, errors.New("无效成绩: " + tokens[i])
		}
		v := result.TimeParserS2F(tokens[i])
		if v == result.UNT || math.IsNaN(v) || math.IsInf(v, 0) || (v <= 0 && v != result.DNF && v != result.DNS) {
			return nil, errors.New("无效成绩: " + tokens[i])
		}
		values = append(values, v)
	}
	if len(values) > m.Rounds {
		return nil, errors.New("成绩数量超过赛制限制")
	}
	return values, nil
}
