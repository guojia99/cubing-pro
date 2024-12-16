package plugin

import (
	"encoding/json"
	"github.com/donnie4w/go-logger/logger"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/robot/types"
	"math/rand"
	"os"
	"path"
	"slices"
	"strings"
)

const MetaFileName = "meta.json"
const MetaImageFileName = "img"

type personMeta struct {
	Enable bool `json:"enable"`

	QuotesKeys []string `json:"quotesKeys"` // 语录触发键
	ImagesKey  []string `json:"imagesKey"`  // 图片触发键

	ImageQuotes []string `json:"imageQuotes"` // 图片附言
	QuotesPath  string   `json:"quotesPath"`  // 语录地址

	Groups []int64 `json:"groups"` // 群ID列表

	imageFiles []string // 文件列表
	quotes     []string // 语录列表
}

func (p *personMeta) rangeImage() (img string, q string) {
	if len(p.imageFiles) == 0 || len(p.ImageQuotes) == 0 {
		return "", ""
	}
	return p.imageFiles[rand.Intn(len(p.imageFiles))], p.ImageQuotes[rand.Intn(len(p.ImageQuotes))]
}

func (p *personMeta) rangeQuote() string {
	if len(p.quotes) == 0 {
		return ""
	}
	return p.quotes[rand.Intn(len(p.quotes))]
}

type PersonValPlugin struct {
	Svc *svc.Svc

	metaImageKeys  map[string]*personMeta
	metaQuotesKeys map[string]*personMeta
}

func getPersonValue(file string) []string {
	var caoGodLine []string
	data, err := os.ReadFile(file)
	if err != nil {
		return caoGodLine
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, "==") {
			continue
		}
		if len(line) == 0 {
			continue
		}
		caoGodLine = append(caoGodLine, line)
	}
	return caoGodLine
}

func (p *PersonValPlugin) init() {
	// todo 冲突处理
	dirs, err := os.ReadDir(p.Svc.Cfg.Robot.PersonValPath)
	if err != nil {
		logger.Errorf("[Robot][Person] 无法读取人物列表")
		return
	}

	p.metaImageKeys = make(map[string]*personMeta)
	p.metaQuotesKeys = make(map[string]*personMeta)

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		metaFile := path.Join(p.Svc.Cfg.Robot.PersonValPath, dir.Name(), MetaFileName)

		data, err := os.ReadFile(metaFile)
		if err != nil {
			logger.Errorf("[Robot][Person] 找不到文件%s", metaFile)
			continue
		}
		var meta personMeta
		if err = json.Unmarshal(data, &meta); err != nil {
			logger.Errorf("[Robot][Person] 文件解析错误%s", metaFile)
			continue
		}

		if !meta.Enable {
			continue
		}

		imageDirPath := path.Join(p.Svc.Cfg.Robot.PersonValPath, dir.Name(), MetaImageFileName)
		imageDir, err := os.ReadDir(imageDirPath)
		if err != nil {
			logger.Errorf("[Robot][Person] 读取图片地址解析错误%s", imageDirPath)
			continue
		}
		for _, file := range imageDir {
			ext := path.Ext(file.Name())
			switch ext {
			case ".jpg", "jpg", ".jpeg", "jpeg", ".png", "png", ".gif", "gif":
			default:
				continue
			}
			meta.imageFiles = append(meta.imageFiles, path.Join(imageDirPath, file.Name()))
		}
		if meta.QuotesPath != "" {
			meta.quotes = getPersonValue(path.Join(p.Svc.Cfg.Robot.PersonValPath, dir.Name(), meta.QuotesPath))
		}

		// 生成key
		for _, key := range meta.ImagesKey {
			p.metaImageKeys[key] = &meta
		}
		for _, key := range meta.QuotesKeys {
			p.metaQuotesKeys[key] = &meta
		}
	}
}

func (p *PersonValPlugin) ID() []string {
	p.init()
	var out []string
	for key := range p.metaImageKeys {
		out = append(out, key)
	}
	for key := range p.metaQuotesKeys {
		out = append(out, key)
	}
	return out
}
func (p *PersonValPlugin) Help() string { return "探索吧" }

func (p *PersonValPlugin) Do(message types.InMessage) (*types.OutMessage, error) {

	// todo 合并 图和语录
	if metaQ, has := p.metaQuotesKeys[message.Message]; has {
		if q := metaQ.rangeQuote(); len(q) > 0 {
			if len(metaQ.Groups) == 0 || slices.Contains(metaQ.Groups, message.GroupID) {
				return message.NewOutMessage(q), nil
			}
		}
	}
	if metaI, has := p.metaImageKeys[message.Message]; has {
		if i, q := metaI.rangeImage(); len(q) > 0 {
			if len(metaI.Groups) == 0 || slices.Contains(metaI.Groups, message.GroupID) {
				return message.NewOutMessageWithImage(q, i), nil
			}
		}
	}
	return nil, nil
}
