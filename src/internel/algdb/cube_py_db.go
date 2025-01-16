package algdb

import (
	"path"
	"strings"

	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type CubePy struct {
	l4eRawData Cube
	l4e        CubeAlgDb
}

func NewCubePy(dbPath string) *CubePy {
	c := &CubePy{}

	_ = utils.ReadJson(path.Join(dbPath, "py", "l4e.json"), &c.l4eRawData)
	c.l4e = c.l4eRawData.ToCubeAlgDb()
	return c
}

func (c *CubePy) ID() []string { return []string{"py", "金字塔", "l4e"} }
func (c *CubePy) Cases() []string {
	return []string{"l4e"}
}
func (c *CubePy) UpdateCases() []string { return []string{} }
func (c *CubePy) UpdateConfig(caseInput string, oldConfig interface{}) (config string, err error) {
	return "", nil
}
func (c *CubePy) Help() string {
	return `金字塔四棱公式查询
a. l4e 查询公式列表
b. l4e S 查询公式`
}
func (c *CubePy) BaseConfig() interface{} {
	return []string{
		"Last Layer",
		"Last 3 Edges",
		"Flipped Edges",
		"Polish Flip",
		"Seperated Bar",
		"Connected Bar",
		"No Bar",
	}
}
func (c *CubePy) Select(selectInput string, config interface{}) (output string, image string, err error) {
	msg := strings.TrimSpace(utils.ReplaceAll(selectInput, "", c.ID()...))

	sp := utils.Split(msg, " ")
	if len(sp) == 0 {
		return c.l4eRawData.CaseList(), "", nil
	}

	if len(sp) != 1 {
		return c.Help(), "", nil
	}

	l := strings.ToLower(sp[0])

	var alg CubeAlg
	for _, a := range c.l4eRawData.AlgInfos {
		if strings.ToLower(a.Name) == l {
			alg = a
			break
		}
	}
	if alg.Name == "" {
		return "case不存在", "", nil
	}

	out, img := alg.Data(c.l4e.Image)
	return out, img, nil
}
