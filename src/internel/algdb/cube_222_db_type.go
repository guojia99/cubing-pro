package algdb

import (
	"fmt"
	"strings"
)

type Cube222Config struct {
	EG EgConfig
}

type (
	Cube222EgAlg struct {
		Alg      []string `json:"alg"`
		Name     string   `json:"name"`
		Group    string   `json:"group"`
		Set      string   `json:"set"`
		Scramble string   `json:"scramble"`
	}

	Cube222Eg struct {
		Keys     map[string][]int        `json:"keys"`
		Sets     map[string][]string     `json:"sets"`
		AlgInfos map[string]Cube222EgAlg `json:"algInfos"`
		Images   map[string]string       `json:"images"`
	}

	CubeEgAlgDb struct {
		Alg   map[string]Cube222EgAlg // key: Set-Group-Name 例如 CLL-U-U5
		Image map[string]string       // svg
	}

	EgConfig struct {
		Group map[string]string `json:"group"`
		Cases map[string]string `json:"cases"`
		Name  map[string]string `json:"name"`
	}
)

func (c Cube222Eg) ToCubeEgAlgDb() CubeEgAlgDb {
	var out = CubeEgAlgDb{
		Alg:   make(map[string]Cube222EgAlg),
		Image: make(map[string]string),
	}

	for k, v := range c.AlgInfos {
		key := fmt.Sprintf("%s_%s", strings.ToLower(v.Set), strings.ToLower(v.Name)) // CLL S3
		out.Alg[key] = v
		out.Image[key] = c.Images[k]
	}

	return out
}
