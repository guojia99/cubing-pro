package algdb

import (
	"fmt"
	"math"
	"path"
	"strings"

	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type Cube222 struct {
	egRawData Cube
	eg        CubeAlgDb
}

func (c *Cube222) ID() []string { return []string{"222", "二阶", "2"} }
func (c *Cube222) Cases() []string {
	return []string{
		"eg", "EG", "EG1", "EG2", "cll", "LEG",
	}
}
func (c *Cube222) UpdateCases() []string { return []string{} }
func (c *Cube222) UpdateConfig(caseInput string, oldConfig interface{}) (config string, err error) {
	return "", nil
}
func (c *Cube222) BaseConfig() interface{} {
	return Cube222Config{
		EG: EgConfig{
			Set: map[string]string{
				"cll":   "CLL",
				"eg0":   "CLL",
				"eg-0":  "CLL",
				"eg1":   "EG-1",
				"eg-1":  "EG-1",
				"eg2":   "EG-2",
				"eg-2":  "EG-2",
				"leg":   "LEG-1",
				"leg-1": "LEG-1",
			},
			Group: map[string]string{
				"s":         "Sune",
				"sune":      "Sune",
				"as":        "Anti-Sune",
				"anti-sune": "Anti-Sune",
				"pi":        "Pi",
				"p":         "Pi",
				"u":         "U",
				"l":         "L",
				"t":         "T",
				"h":         "H",
			},
			Name: map[string]string{
				"Sune":      "S",
				"Anti-Sune": "As",
				"Pi":        "Pi",
				"U":         "U",
				"L":         "L",
				"T":         "T",
				"H":         "H",
			},
		},
	}
}

func (c *Cube222) Help() string {
	return `二阶公式查询
EG:
a. 222 eg1 s1 查询具体公式及图片展示
b. 目前可查询: cll\eg0 eg1 eg2 leg
c. eg case有: S, As, Pi, L, T, U, H
`
}

func (c *Cube222) Select(selectInput string, config interface{}) (output string, image string, err error) {
	msg := strings.TrimSpace(utils.ReplaceAll(selectInput, "", c.ID()...))
	sp := utils.Split(msg, " ")

	if len(sp) == 0 {
		return c.Help(), "", nil
	}
	Case := strings.ToLower(sp[0])
	switch Case {
	case "eg1", "eg2", "cll", "eg0", "eg-1", "eg-2", "eg-0", "leg":
		return c.selectEg(sp, config)
	}
	return c.Help(), "", nil
}

func NewCube222(dbPath string) *Cube222 {
	b := &Cube222{}
	_ = utils.ReadJson(path.Join(dbPath, "222", "eg.json"), &b.egRawData)

	b.eg = b.egRawData.ToCubeAlgDb() // 简化数据结构，统一数据
	return b
}

func (c *Cube222) selectEg(selectInput []string, config interface{}) (output string, image string, err error) {
	if config == nil {
		config = c.BaseConfig()
	}
	cfg := config.(Cube222Config)

	if len(selectInput) != 2 {
		return c.Help(), "", nil
	}

	setStr, nameStr := strings.ToLower(selectInput[0]), strings.ToLower(selectInput[1])
	set, ok := cfg.EG.Set[setStr] // cll
	if !ok {
		return c.Help(), "", nil
	}

	num := utils.GetNum(nameStr)
	if math.IsNaN(num) || num <= 0 || num > 6 {
		return c.Help(), "", nil
	}
	groupStr := utils.ReplaceAll(nameStr, "", fmt.Sprintf("%d", int(num)))
	group, ok := cfg.EG.Group[groupStr]
	if !ok {
		return c.Help(), "", nil
	}

	key := fmt.Sprintf("%s_%s_%s", strings.ToLower(set), strings.ToLower(group), nameStr) // set group name
	alg, ok := c.eg.Alg[key]
	if !ok {
		return fmt.Sprintf("找不到该case: %s", key), "", nil
	}

	out, img := alg.Data(c.eg.Image)
	return out, img, nil
}
