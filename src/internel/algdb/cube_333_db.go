package algdb

import (
	"path"
	"strings"

	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type Cube333 struct {
	pllRawData   Cube
	pll          CubeAlgDb
	ohPllRawData Cube
	ohPll        CubeAlgDb
}

func (c *Cube333) ID() []string { return []string{"333", "三阶"} }
func (c *Cube333) Cases() []string {
	return []string{
		"oll", "o",
		"pll", "p",
		"cmll", "cm",
		//"zbll", "zb",
	}
}

func (c *Cube333) UpdateCases() []string { return []string{} }
func (c *Cube333) UpdateConfig(caseInput string, oldConfig interface{}) (config string, err error) {
	return "", nil
}
func (c *Cube333) BaseConfig() interface{} {
	return map[string]interface{}{}
}

func (c *Cube333) Help() string {
	return `三阶公式查询
PLL: pll Ja
CMLL: 待更新
ZBLL: 待更新
`
}
func (c *Cube333) Select(selectInput string, config interface{}) (output string, image string, err error) {
	msg := strings.TrimSpace(utils.ReplaceAll(selectInput, "", c.ID()...))
	sp := utils.Split(msg, " ")

	if len(sp) == 0 {
		return c.Help(), "", nil
	}
	Case := strings.ToLower(sp[0])
	switch Case {
	case "pll", "p":
		return c.selectPll(sp, config)
	}

	return "", "", nil
}

func NewCube333(dbPath string) *Cube333 {
	b := &Cube333{}
	_ = utils.ReadJson(path.Join(dbPath, "333", "pll.json"), &b.pllRawData)
	_ = utils.ReadJson(path.Join(dbPath, "333", "oh-pll.json"), &b.ohPllRawData)
	b.pll = b.pllRawData.ToCubeAlgDb() // 简化数据结构，统一数据
	b.ohPll = b.ohPllRawData.ToCubeAlgDb()
	return b
}

func (c *Cube333) selectPll(selectInput []string, config interface{}) (output string, image string, err error) {
	if config == nil {
		config = c.BaseConfig()
	}
	if len(selectInput) != 2 {
		return c.Help(), "", err
	}

	return "", "", err
}
