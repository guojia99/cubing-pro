package algdb

import (
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path"
	"strings"
	"time"

	"github.com/fogleman/gg"

	"github.com/2mf8/Better-Bot-Go/log"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type SQ1CspDB struct {
	DBPath    string
	ImagePath string
	TempPath  string
	FontTTf   string

	data     cspAlgMap
	dataList []string
}

func NewSQ1CspDB(dbPath string, imagePath string, tmpPath string, FontTTf string) *SQ1CspDB {
	s := &SQ1CspDB{
		ImagePath: imagePath,
		DBPath:    dbPath,
		TempPath:  tmpPath,
		FontTTf:   FontTTf,
	}
	s.init()
	return s
}

func (s *SQ1CspDB) init() {
	_ = os.MkdirAll(s.TempPath, os.ModePerm)

	file, err := os.ReadFile(s.DBPath)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	var in map[string]interface{}
	err = json.Unmarshal(file, &in)
	if err != nil {
		return
	}

	b0, _ := json.Marshal(in[listKey])
	_ = json.Unmarshal(b0, &s.dataList)

	delete(in, listKey)

	b, _ := json.Marshal(in)
	_ = json.Unmarshal(b, &s.data)

}

func (s *SQ1CspDB) ID() []string          { return []string{"csp", "CSP", "Csp", "吃薯片"} }
func (s *SQ1CspDB) Cases() []string       { return []string{} }
func (s *SQ1CspDB) UpdateCases() []string { return []string{"修改配置"} }

func (s *SQ1CspDB) Help() string { return "'csp 桶-桶' : 可获取对应公式 " }

func (s *SQ1CspDB) Select(selectInput string, config interface{}) (output string, image string, err error) {
	if config == nil {
		config = s.BaseConfig()
	}

	fmt.Println(utils.ReplaceAll(selectInput, "", " "), "==========")
	if utils.ReplaceAll(selectInput, "", " ") == "吃薯片" {
		return "嘎崩脆", "", nil
	}

	selectInput = utils.ReplaceAll(selectInput, "", s.ID()...)
	var input []string

	if strings.Contains(selectInput, "/") {
		input = strings.Split(selectInput, "/")
	} else {
		xxInput := strings.Split(selectInput, " ")
		for _, x := range xxInput {
			if len(x) > 0 {
				input = append(input, x)
			}
		}
	}
	if len(input) < 2 {
		return "", "", fmt.Errorf("格式应当为: 'csp 桶 桶' 或 'csp star/star'")
	}
	reConfig := s.reConfig(config.(map[string]string))

	key1, ok1 := reConfig[utils.ReplaceAll(input[0], "", " ")]
	key2, ok2 := reConfig[utils.ReplaceAll(input[1], "", " ")]
	if !ok1 || !ok2 {
		out := s.getList(config.(map[string]string))
		return "", "", fmt.Errorf("`%s`, `%s`的配置名称不存在\n 请参考\n %s\n", input[0], input[1], out)
	}

	data, algKey, err := s.getData(key1, key2)
	if err != nil {
		return "", "", err
	}
	out := fmt.Sprintf("形态 ====> %s\n", algKey)
	base, baseOk := data[baseKey]
	if baseOk {
		out += fmt.Sprintf("-- 基础\n")
		out += fmt.Sprintf("a.偶(%d) %s\n", strings.Count(base.Even, "/"), base.Even)
		out += fmt.Sprintf("b.奇(%d) %s\n", strings.Count(base.Odd, "/"), base.Odd)
	}
	mirror, mirrorOk := data[mirrorKey]
	if mirrorOk {
		out += fmt.Sprintf("-- 倒置\n")
		out += fmt.Sprintf("a.偶(%d) %s\n", strings.Count(mirror.Even, "/"), mirror.Even)
		out += fmt.Sprintf("b.奇(%d) %s\n", strings.Count(mirror.Odd, "/"), mirror.Odd)
	}

	// 合并图片
	img := s.getImage(base, mirror)

	fmt.Println(img)
	return out, img, nil
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

func (s *SQ1CspDB) getImage(base, mirror cspAlg) string {
	bIEv, bIOd := "", ""
	mIEv, mIOd := "", ""

	if len(base.Image) == 2 {
		bIEv, bIOd = base.Image[0], base.Image[1]
	}
	if len(mirror.Image) == 2 {
		mIEv, mIOd = mirror.Image[0], mirror.Image[1]
	}

	if len(bIEv) == 0 || len(bIOd) == 0 {
		return ""
	}
	bIEv = path.Join(s.ImagePath, bIEv)
	bIOd = path.Join(s.ImagePath, bIOd)
	mIEv = path.Join(s.ImagePath, mIEv)
	mIOd = path.Join(s.ImagePath, mIOd)

	bIEvImg, err1 := utils.OpenImage(bIEv)
	bIOdImg, err2 := utils.OpenImage(bIOd)
	if err1 != nil || err2 != nil {
		return ""
	}

	width := bIEvImg.Bounds().Dx()
	height := bIEvImg.Bounds().Dy()

	const imageMesDy = 50
	newWidth := bIEvImg.Bounds().Dx()
	newHeight := bIEvImg.Bounds().Dy() * 2

	mIEvImg, err3 := utils.OpenImage(mIEv)
	mIOdImg, err4 := utils.OpenImage(mIOd)
	if err3 == nil && err4 == nil {
		newWidth *= 2
	}
	newImage := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	draw.Draw(newImage, image.Rect(0, 0, width, height), bIEvImg, image.Point{}, draw.Over)
	draw.Draw(newImage, image.Rect(0, height, width, newHeight), bIOdImg, image.Point{}, draw.Over)

	if err3 == nil && err4 == nil {
		draw.Draw(newImage, image.Rect(width, 0, newWidth, height), mIEvImg, image.Point{}, draw.Over)
		draw.Draw(newImage, image.Rect(width, height, newWidth, newHeight), mIOdImg, image.Point{}, draw.Over)
	}
	outputPath := path.Join(s.TempPath, fmt.Sprintf("%d.png", time.Now().UnixNano()))
	//outputPath := path.Join(s.TempPath, "test.png")
	// 写入文字
	dc := gg.NewContextForImage(newImage)
	if err := dc.LoadFontFace(s.FontTTf, 30); err != nil {
		fmt.Printf("无法加载字体: %v\n", err)
		return ""
	}
	dc.SetRGB(206, 27, 183)
	dc.DrawStringAnchored("Base", float64(40), float64(newHeight-20), 0.5, 0.5)
	if err3 == nil && err4 == nil {
		dc.SetRGB(181, 224, 133)
		dc.DrawStringAnchored("Invert", float64(newWidth-50), float64(newHeight-20), 0.5, 0.5)
	}

	err := utils.SaveImage(outputPath, dc.Image())
	if err != nil {
		return ""
	}
	return outputPath
}

func (s *SQ1CspDB) getList(config map[string]string) (out string) {

	idx := 1
	for _, key := range s.dataList {
		st := strings.Split(key, "/")
		k1, k2 := strings.TrimRight(st[0], " "), strings.TrimLeft(st[1], " ")
		if t, ok := config[k1]; ok {
			k1 = t
		}
		if t, ok := config[k2]; ok {
			k2 = t
		}

		out += fmt.Sprintf("%d. %s / %s\n", idx, k1, k2)
		idx++
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
		"5-3":        "5-3",
		"square":     "方",
		"kite":       "筝",
		"scallop":    "贝",
		"shield":     "盾",
		"barrel":     "桶",
		"muffin":     "菇",
		"fist":       "拳",
		"left fist":  "左拳",
		"right fist": "右拳",
		"pawn":       "爪",
		"left paw":   "左爪",
		"right paw":  "右爪",
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
	Even  string   `json:"even"`
	Odd   string   `json:"odd"`
	Image []string `json:"image"`
}

const (
	baseKey   = "base"
	mirrorKey = "mirror"
	listKey   = "___list"
)

type cspAlgMap map[string]map[string]cspAlg // map[case] map[base|mirror] cspAlg ||| __ []string
