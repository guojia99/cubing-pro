package algdb

import (
	"fmt"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/guojia99/cubing-pro/src/internel/algdb/script"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type bldAlg map[string][]string

// map 最外层为三循环
// [] 为不同的公式
// [2][]string
//    第一层为公式列表
//    第二层为公式使用者

type bldManMadeAlg map[string][][2][]string

const (
	edgeAlgToStandard = "edge/edgeAlgToStandard.json"
	edgeCodeToPos     = "edge/edgeCodeToPos.json"
	edgePosToCode     = "edge/edgePosToCode.json"
	edgeManMade       = "edge/edgeAlgToInfoManmade.json"

	cornerAlgToStandard = "corner/cornerAlgToStandard.json"
	cornerCodeToPos     = "corner/cornerCodeToPos.json"
	cornerPosToCode     = "corner/cornerPosToCode.json"
	cornerManMade       = "corner/cornerAlgToInfoManmade.json"

	commutatorJs = "commutator.js"
)

type BldDB struct {
	bldPath string

	cornerAlgToStandard, cornerCodeToPos, cornerPosToCode map[string]string // 查询映射表
	edgeAlgToStandard, edgeCodeToPos, edgePosToCode       map[string]string // 查询映射表

	edgeManMade   bldManMadeAlg
	cornerManMade bldManMadeAlg

	edgeInfo   map[string]bldAlg
	cornerInfo map[string]bldAlg
}

func NewBldDB(bldPath string) *BldDB {
	script.InitCommutator(path.Join(bldPath, commutatorJs))

	b := &BldDB{
		bldPath:    bldPath,
		edgeInfo:   make(map[string]bldAlg),
		cornerInfo: make(map[string]bldAlg),
	}

	// edge 映射表
	_ = utils.ReadJson(path.Join(bldPath, edgeAlgToStandard), &b.edgeAlgToStandard)
	_ = utils.ReadJson(path.Join(bldPath, edgeCodeToPos), &b.edgeCodeToPos)
	_ = utils.ReadJson(path.Join(bldPath, edgePosToCode), &b.edgePosToCode)

	// corner 映射表
	_ = utils.ReadJson(path.Join(bldPath, cornerAlgToStandard), &b.cornerAlgToStandard)
	_ = utils.ReadJson(path.Join(bldPath, cornerCodeToPos), &b.cornerCodeToPos)
	_ = utils.ReadJson(path.Join(bldPath, cornerPosToCode), &b.cornerPosToCode)

	// 公式表
	_ = utils.ReadJson(path.Join(bldPath, edgeManMade), &b.edgeManMade)
	_ = utils.ReadJson(path.Join(bldPath, cornerManMade), &b.cornerManMade)

	// 独立公式表
	edgeInfos := map[string][]string{
		"edge/edgeAlgToInfo.json": {"info", "噩梦"},
	}
	for key, p := range edgeInfos {
		var info = make(bldAlg)
		_ = utils.ReadJson(path.Join(bldPath, key), &info)
		for _, v := range p {
			b.edgeInfo[v] = info
		}
	}
	cornerInfos := map[string][]string{
		"corner/cornerAlgToInfo.json":        {"info", "噩梦"},
		"corner/cornerAlgToInfoBalance.json": {"balance", "平衡"},
		"corner/cornerAlgToInfoYuanzi.json":  {"圆子", "yuanzi"},
	}
	for key, p := range cornerInfos {
		var info = make(bldAlg)
		_ = utils.ReadJson(path.Join(bldPath, key), &info)
		for _, v := range p {
			b.cornerInfo[v] = info
		}
	}

	return b
}

func (b *BldDB) ID() []string          { return []string{"bld", "3bf", "三蝙蝠"} }
func (b *BldDB) Cases() []string       { return []string{"edge", "e", "E", "棱", "corner", "c", "C", "角"} }
func (b *BldDB) UpdateCases() []string { return nil }
func (b *BldDB) Help() string {
	return `数据来源: blddb|王子兴
a. bld 棱|角 ADE
b.bld 角[人造|噩梦|圆子|平衡] ADG
c.bld 角[man|info|yuanzi|balance] ADG
`
}

var classInfo = []string{
	"人造", "噩梦", "圆子", "平衡", "man", "info", "yuanzi", "balance",
}

func (b *BldDB) Select(selectInput string, config interface{}) (output string, image string, err error) {
	msg := strings.TrimSpace(utils.ReplaceAll(selectInput, "", b.ID()...))
	sp := strings.Split(msg, " ")
	if len(sp) != 2 {
		return "", "", fmt.Errorf(b.Help())
	}

	class, result := sp[0], strings.ToUpper(sp[1])
	cla := "人造"
	if matches := regexp.MustCompile(`\[(.*?)\]`).FindStringSubmatch(class); len(matches) >= 2 {
		cla = matches[1]
		class = utils.ReplaceAll(class, "", matches[0])
	}

	if !slices.Contains(classInfo, cla) {
		return "", "", fmt.Errorf("不存在该类型 `%s`\n请在以下类型中选择: %+v\n", cla, classInfo)
	}

	// 获取解析case
	sCase, ok := "", false
	isEdge, isCorner := false, false
	switch strings.ToLower(class) {
	case "e", "edge", "棱":
		isEdge = true
		// 将result转换为制定格式
		sCase, ok = b.edgeAlgToStandard[b.updateEdgeResult(result)]
	case "c", "corner", "角":
		isCorner = true
		sCase, ok = b.cornerAlgToStandard[b.updateCornerResult(result)]
	default:
		return "", "", fmt.Errorf("请选择 棱或角")
	}
	if !ok {
		return "", "", fmt.Errorf("找不到该case `%s`", sp[1])
	}

	// 渲染公式
	out := fmt.Sprintf("%s %s Case %s ==> %s\n", class, cla, result, sCase)
	if slices.Contains([]string{"人造", "man"}, cla) {
		var getData bldManMadeAlg
		if isEdge {
			getData = b.edgeManMade
		}
		if isCorner {
			getData = b.cornerManMade
		}
		if getData == nil || len(getData) == 0 {
			return "", "", fmt.Errorf("暂无数据")
		}

		data, ok2 := getData[sCase]
		if !ok2 {
			return "", "", fmt.Errorf("找不到该case `%s`", sp[1])
		}
		for idx, res := range data {
			out += fmt.Sprintf("%d. (%d)\t", idx+1, len(res[1]))
			for _, alg := range res[0] {
				comm, errx := script.Commutator(alg)
				if errx != nil {
					out += fmt.Sprintf("公式: %s\n", alg)
				} else {
					out += fmt.Sprintf("公式: %s\n\t\t交换子:%s\n", alg, comm)
				}
				out += "\n"
			}
		}
	} else {
		// 其他info
		var getData bldAlg
		var ok3 bool
		if isEdge {
			getData, ok3 = b.edgeInfo[cla]
		}
		if isCorner {
			getData, ok3 = b.cornerInfo[cla]
		}
		if !ok3 {
			return "", "", fmt.Errorf("找不到公式库")
		}

		data, ok4 := getData[sCase]
		if !ok4 {
			return "", "", fmt.Errorf("找不到该case `%s`", sp[1])
		}

		maxIdx := 10
		for idx, alg := range data {
			if idx >= maxIdx {
				continue
			}
			comm, err := script.Commutator(alg)
			if err == nil {
				out += fmt.Sprintf("%d. 公式: %s\n交换子: %s\n", idx+1, alg, comm)
			} else {
				out += fmt.Sprintf("%d. 公式: %s\n", idx+1, alg)
			}

		}
	}
	return out, "", err
}

func (b *BldDB) UpdateConfig(caseInput string, oldConfig interface{}) (config string, err error) {
	return "", nil
}
func (b *BldDB) BaseConfig() interface{} {
	mp := make(map[string]map[string]string)
	return mp
}

func (b *BldDB) updateEdgeResult(res string) string {
	if strings.Index(res, "-") == -1 {
		return res
	}
	sp := strings.Split(res, "-")
	if len(sp) != 3 {
		return res
	}
	out := ""
	for _, v := range sp {
		x, ok := b.edgePosToCode[v]
		if ok {
			out += x
		}
	}
	return out
}

func (b *BldDB) updateCornerResult(res string) string {
	if strings.Index(res, "-") == -1 {
		return res
	}
	sp := strings.Split(res, "-")
	if len(sp) != 3 {
		return res
	}
	out := ""
	for _, v := range sp {
		if len(v) != 3 {
			continue
		}
		x, ok := b.cornerPosToCode[v]
		if ok {
			out += x
			continue
		}

		v2 := fmt.Sprintf("%s%s%s", v[0:1], v[2:], v[1:2])
		y, ok := b.cornerPosToCode[v2]
		if ok {
			out += y
		}
	}

	if len(out) != 3 {
		return res
	}

	if _, ok := b.cornerAlgToStandard[out]; ok {
		return out
	}

	return res
}
