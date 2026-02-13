package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/guojia99/cubing-pro/src/internel/algs"
)

// import json
//
//	data = {
//	   "name": "3x3-F2L-Trainer",
//	   "set_keys": [
//	       "Front Right", "Front Left", "Back Left", "Back Right"
//	   ]
//	}
//
// ALG_INFO = "algs_info.json"
// ALG_SET_INFO = "algsets_info.json"
// COMBINED = "combined.json"
// GROUPS_INFO = "groups_info.json"
// SCRAMBLES = "scrambles.json"
func deduplicate(groups []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0)

	for _, g := range groups {
		if _, exists := seen[g]; !exists {
			seen[g] = struct{}{}
			result = append(result, g)
		}
	}

	return result
}

type TrData struct {
	Name    string
	SetKeys []string
}

var data = TrData{
	Name: "./SQ1-EO-Trainer",
	SetKeys: []string{
		"EO",
	},
}

type OutputJSON struct {
	Index string              `json:"index"`
	Name  string              `json:"name"`
	Algs  map[string][]string `json:"algs"`
	Setup string              `json:"setup"`
	Group string              `json:"group"`
	Image string              `json:"image"`
}

func writeToJson(file string, data interface{}) {
	f, _ := os.Create(file)
	defer f.Close()
	d, _ := json.MarshalIndent(data, "", " ")
	_, _ = f.WriteString(string(d))
}

const (
	algsInfoFile         = "algs_info.json"
	algsetsInfoFile      = "algsets_info.json"
	algImagesInfoFile    = "combined.json"
	algGroupFile         = "groups_info.json"
	algScramblesInfoFile = "scrambles.json"
)

// 反转单个魔方操作的方向
func reverseMove(move string) string {
	if strings.HasSuffix(move, "2") {
		return move // 2表示180度，不变
	} else if strings.HasSuffix(move, "'") {
		return move[:len(move)-1] // 去掉 '，变顺时针
	} else {
		return move + "'" // 加上 '，变逆时针
	}
}

// 反转整段打乱公式
func reverseScramble(scramble string) string {
	moves := strings.Fields(strings.TrimSpace(scramble))

	// 反转顺序
	for i, j := 0, len(moves)-1; i < j; i, j = i+1, j-1 {
		moves[i], moves[j] = moves[j], moves[i]
	}

	// 反转每一步方向
	for i := range moves {
		moves[i] = reverseMove(moves[i])
	}

	return strings.Join(moves, " ")
}

// 计算公式长度（按 move 数量）
func formulaLength(s string) int {
	return len(strings.Fields(strings.TrimSpace(s)))
}

// 找出最短公式
func findShortestFormula(formulas []string) string {
	if len(formulas) == 0 {
		return ""
	}

	shortest := formulas[0]
	minLen := formulaLength(shortest)

	for _, f := range formulas[1:] {
		l := formulaLength(f)
		if l < minLen {
			minLen = l
			shortest = f
		}
	}

	return shortest
}

func parseData(output []OutputJSON) *algs.AlgorithmConfigWithTrainer {
	for idx, _ := range output {
		output[idx].Index = fmt.Sprintf("%d", idx+1)
	}

	_ = os.MkdirAll(data.Name, 0755)

	out := &algs.AlgorithmConfigWithTrainer{
		AlgsInfo:    make(map[string]algs.TrainerAlgorithm),
		AlgsetsInfo: make(map[string][]string),
		GroupsInfo:  make(map[string][]int),
		Images:      make(map[string]string),
		Scrambles:   make(map[string][]string),
	}

	for index, in := range output {
		for set, alg := range in.Algs {
			trainerAlgorithm := algs.TrainerAlgorithm{
				Algs:     alg,
				Name:     in.Name,
				Group:    in.Group,
				Algset:   set,
				Scramble: in.Setup,
			}
			key := fmt.Sprintf("%d", index+1)
			out.AlgsInfo[key] = trainerAlgorithm

			// 打乱
			out.Scrambles[key] = []string{in.Setup}
			if in.Setup == "" {
				out.Scrambles[key] = []string{findShortestFormula(alg)}
			}

			// 图片
			out.Images[key] = in.Image

			// 分组
			groupKey := fmt.Sprintf("%s %s", set, in.Group)
			if _, ok := out.GroupsInfo[groupKey]; !ok {
				out.GroupsInfo[groupKey] = make([]int, 0)
			}
			out.GroupsInfo[groupKey] = append(out.GroupsInfo[groupKey], index+1)

			// set
			if _, ok := out.AlgsetsInfo[set]; !ok {
				out.AlgsetsInfo[set] = make([]string, 0)
			}
			out.AlgsetsInfo[set] = append(out.AlgsetsInfo[set], groupKey)
		}
	}

	for key := range out.AlgsetsInfo {
		out.AlgsetsInfo[key] = deduplicate(out.AlgsetsInfo[key])
	}

	writeToJson(path.Join(data.Name, algsInfoFile), out.AlgsInfo)
	writeToJson(path.Join(data.Name, algImagesInfoFile), out.Images)
	writeToJson(path.Join(data.Name, algGroupFile), out.GroupsInfo)
	writeToJson(path.Join(data.Name, algScramblesInfoFile), out.Scrambles)
	writeToJson(path.Join(data.Name, algsetsInfoFile), out.AlgsetsInfo)

	//
	//for idx, i := range in {
	//	// 打乱
	//	out.Scrambles[i.Index] = []string{i.Setup}
	//
	//	if _, ok := out.GroupsInfo[i.Group]; !ok {
	//		out.GroupsInfo[i.Group] = make([]int, 0)
	//	}
	//	// 分组
	//	out.GroupsInfo[i.Group] = append(out.GroupsInfo[i.Group], idx+1)
	//	// 图片
	//	out.Images[i.Index] = i.Image
	//
	//}

	return out
}

func main() {
	var output []OutputJSON
	d, _ := os.ReadFile("./output.json")
	json.Unmarshal(d, &output)

	parseData(output)
}
