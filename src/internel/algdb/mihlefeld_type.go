package algdb

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type Cube222Config struct {
	EG   EgConfig
	TCll map[string]string
}

type (
	CubeAlg struct {
		Alg      []string `json:"a"`
		Name     string   `json:"name"`
		Group    string   `json:"group"`
		Set      string   `json:"algset"`
		Scramble string   `json:"s"`
	}

	Cube struct {
		Keys     map[string][]int    `json:"keys"`
		Sets     map[string][]string `json:"sets"`
		AlgInfos map[string]CubeAlg  `json:"algInfos"`
		Images   map[string]string   `json:"images"`
	}

	CubeAlgDb struct {
		Alg   map[string]CubeAlg // key: Set-Group-Name 例如 CLL-U-U5
		Image map[string]string  // svg
	}

	EgConfig struct {
		Set   map[string]string
		Group map[string]string
		Name  map[string]string
	}
)

func (c Cube) ToCubeAlgDb() CubeAlgDb {
	var out = CubeAlgDb{
		Alg:   make(map[string]CubeAlg),
		Image: make(map[string]string),
	}

	for k, v := range c.AlgInfos {
		key := fmt.Sprintf("%s_%s_%s", strings.ToLower(v.Set), strings.ToLower(v.Group), strings.ToLower(v.Name)) // CLL S3
		//fmt.Println(key)
		out.Alg[key] = v
		out.Image[key] = c.Images[k]
	}

	return out
}

func (c Cube) CaseList() string {
	out := "Case列表\n"
	idx := 1
	for _, val := range c.Sets["L4E"] {
		out += fmt.Sprintf("%d.%s: ", idx, val)
		for _, k := range c.Keys[val] {
			alg := c.AlgInfos[fmt.Sprintf("%d", k)]
			out += fmt.Sprintf("%s、", alg.Name)
		}
		out = strings.TrimRight(out, "、")
		out += "\n"
		idx++
	}
	return out
}

func (a CubeAlg) Data(image map[string]string) (string, string) {
	out := fmt.Sprintf("公式: %s - %s\n", a.Set, a.Name)
	out += fmt.Sprintf("打乱: %s\n", a.Scramble)
	out += "---------------\n"
	for idx, val := range a.Alg {
		out += fmt.Sprintf("%d. %s\n", idx+1, val)
	}
	key := fmt.Sprintf("%s_%s_%s", strings.ToLower(a.Set), strings.ToLower(a.Group), strings.ToLower(a.Name)) // CLL S3
	svgImg, ok := image[key]
	if !ok {
		return out, ""
	}

	imgPath := path.Join("/tmp", fmt.Sprintf("%d.png", time.Now().UnixNano()))
	if err := utils.SaveSvgToImage(svgImg, imgPath); err != nil {
		return out, ""
	}
	return out, imgPath
}
