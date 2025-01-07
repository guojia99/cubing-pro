package algdb

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/2mf8/Better-Bot-Go/log"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type SQ1CspDB struct {
	DBPath string `json:"db_path"`

	data cspAlgMap
}

func NewSQ1CspDB(dbPath string) *SQ1CspDB {
	s := &SQ1CspDB{
		DBPath: dbPath,
	}
	s.init()
	return s
}

func (s *SQ1CspDB) init() {
	file, err := os.ReadFile(s.DBPath)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	err = json.Unmarshal(file, &s.data)
}

func (s *SQ1CspDB) ID() []string          { return []string{"csp", "CSP", "Csp", "吃薯片"} }
func (s *SQ1CspDB) Cases() []string       { return []string{} }
func (s *SQ1CspDB) UpdateCases() []string { return []string{"修改配置"} }

func (s *SQ1CspDB) Help() string { return "'csp 桶-桶' : 可获取对应公式 " }

func (s *SQ1CspDB) Select(selectInput string, config interface{}) (output string, err error) {
	if config == nil {
		config = s.BaseConfig()
	}
	selectInput = utils.ReplaceAll(selectInput, "", s.ID()...)
	var input []string

	if strings.Contains(selectInput, "/") {
		input = strings.Split(selectInput, "/")
	} else {
		input = strings.Split(selectInput, " ")
	}
	if len(input) < 2 {
		return "", fmt.Errorf("格式应当为: 'csp 桶 桶' 或 'csp star/star'")
	}
	reConfig := s.reConfig(config.(map[string]string))

	key1, ok1 := reConfig[utils.ReplaceAll(input[0], "", " ")]
	key2, ok2 := reConfig[utils.ReplaceAll(input[1], "", " ")]
	if !ok1 || !ok2 {
		return "", fmt.Errorf("`%s`, `%s`的配置名称不存在", input[0], input[1])
	}

	data, algKey, err := s.getData(key1, key2)
	if err != nil {
		return "", err
	}
	out := fmt.Sprintf("形态 ====> %s\n", algKey)
	base, baseOk := data[baseKey]
	if baseOk {
		out += fmt.Sprintf("-- 基础\n")
		out += fmt.Sprintf("a.偶(%d) %s\n", len(strings.Split(base.Even, "/")), base.Even)
		out += fmt.Sprintf("b.奇(%d) %s\n", len(strings.Split(base.Odd, "/")), base.Odd)
	}
	mirror, mirrorOk := data[mirrorKey]
	if mirrorOk {
		out += fmt.Sprintf("-- 镜像\n")
		out += fmt.Sprintf("a.偶(%d) %s\n", len(strings.Split(mirror.Even, "/")), mirror.Even)
		out += fmt.Sprintf("b.奇(%d) %s\n", len(strings.Split(mirror.Odd, "/")), mirror.Odd)
	}
	return out, nil
}

func (s *SQ1CspDB) UpdateConfig(updateInput string, oldConfig interface{}) (config string, err error) {
	return
}

func (s *SQ1CspDB) getData(key1, key2 string) (cspAlg map[string]cspAlg, key string, err error) {
	algKey1 := fmt.Sprintf("%s / %s", key1, key2)
	algKey2 := fmt.Sprintf("%s / %s", key2, key1)
	data1, dok1 := s.data[algKey1]
	if dok1 {
		return data1, algKey1, nil
	}
	data2, dok2 := s.data[algKey2]
	if dok2 {
		return data2, algKey2, nil
	}
	return cspAlg, "", fmt.Errorf("找不到该形态 `%s - %s`", key1, key2)
}

func (s *SQ1CspDB) reConfig(mp map[string]string) map[string]string {
	var out = make(map[string]string)
	for k, v := range mp {
		out[v] = k
	}
	return out
}

func (s *SQ1CspDB) BaseConfig() interface{} {
	var mp = map[string]string{
		"star":       "六星",
		"8":          "8",
		"4-4":        "4-4",
		"6-2":        "6-2",
		"7-1":        "7-1",
		"square":     "方",
		"kite":       "筝",
		"scallop":    "贝",
		"shield":     "盾",
		"barrel":     "桶",
		"mushroom":   "菇",
		"fist":       "拳",
		"left fist":  "左拳",
		"right fist": "右拳",
		"pawn":       "爪",
		"left pawn":  "左爪",
		"right pawn": "右爪",
		"pair":       "对",
		"line":       "直线",
		"l":          "拐",
		"6":          "6",
		"5-1":        "5-1",
		"left 5-1":   "左5-1",
		"right 5-1":  "右5-1",
		"4-2":        "4-2",
		"left 4-2":   "左4-2",
		"right 4-2":  "右4-2",
		"4-1-1":      "4-1-1",
		"3-3":        "3-3",
		"3-2-1":      "3-2-1",
		"3-1-2":      "3-1-2",
		"2-2-2":      "2-2-2",
	}
	return mp
}

type cspAlg struct {
	Even string `json:"even"`
	Odd  string `json:"odd"`
}

const (
	baseKey   = "base"
	mirrorKey = "mirror"
)

type cspAlgMap map[string]map[string]cspAlg // map[case] map[base|mirror] cspAlg
