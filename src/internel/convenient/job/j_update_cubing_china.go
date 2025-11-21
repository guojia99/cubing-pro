package job

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/log"
	"gorm.io/gorm"
)

const cubingChinaCompsUrl = "https://cubing.com/competition?year=&type=WCA&province=&event=&lang=zh_cn&page=%d"
const wcaCompsUrl = "https://www.worldcubeassociation.org/competitions/%s"
const UpdateCubingCompetitionKey = "updateCubingCompetitions_job_key"

type CubingCompetition struct {
	ID string `json:"id"` // https://cubing.com/competition/Please-Be-Quiet-Xian-2025 -> Please-Be-Quiet-Xian-2025
	//Url     string `json:"url"`     // https://cubing.com/competition/Please-Be-Quiet-Xian-2025
	//WcaUrl  string `json:"wcaUrl"`  // wca
	WcaID   string `json:"wcaid"`   // 去掉 `-`， PleaseBeQuietXian2025
	Name    string `json:"name"`    // 2025WCA西安安静赛
	EnName  string `json:"enName"`  // 英文名
	City    string `json:"city"`    // 陕西西安
	Address string `json:"address"` // 莲湖区大寨路19号晶鑫商业广场A座3、4号电梯上3楼
}

func parseCubingCompetition(tr *goquery.Selection) (*CubingCompetition, error) {
	// 提取所有 td
	tds := tr.Find("td")
	if tds.Length() < 5 {
		return nil, nil // 或返回错误，根据你的需求
	}

	aTag := tds.Eq(1).Find("a.comp-type-wca")
	url, exists := aTag.Attr("href")
	if !exists {
		return nil, nil
	}

	name := strings.TrimSpace(aTag.Text())

	province := strings.TrimSpace(tds.Eq(2).Text())
	city := strings.TrimSpace(tds.Eq(3).Text())
	address := strings.TrimSpace(tds.Eq(4).Text())

	// 从 URL 中提取 ID: https://cubing.com/competition/Please-Be-Quiet-Xian-2025
	var id string
	if strings.HasPrefix(url, "https://cubing.com/competition/") {
		id = strings.TrimPrefix(url, "https://cubing.com/competition/")
	} else {
		id = "" // 或处理错误
	}

	// 构造 WcaID：移除所有 '-'
	wcaID := regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(id, "")
	enName := regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(id, " ")

	return &CubingCompetition{
		ID: id,
		//Url:     url,
		WcaID: wcaID,
		//WcaUrl:  fmt.Sprintf(wcaCompsUrl, wcaID),
		Name:    name,
		EnName:  enName,
		City:    province + city,
		Address: address,
	}, nil
}

func getCompetitionWithPage(page int) ([]*CubingCompetition, int, error) {
	resp, err := http.Get(fmt.Sprintf(cubingChinaCompsUrl, page))
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, 0, errors.New(resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	// 统计
	var count int
	summary := doc.Find("#yw1 .summary").Text()
	if summary == "" {
		return nil, 0, errors.New("not summary")
	}
	re := regexp.MustCompile(`共\s*(\d+)\s*条`)
	matches := re.FindStringSubmatch(summary)
	if len(matches) < 2 {
		return nil, 0, errors.New("not matches summary")
	}
	count, _ = strconv.Atoi(matches[1])

	// 比赛数据
	var out []*CubingCompetition
	doc.Find("tbody").Each(func(i int, tbody *goquery.Selection) {
		tbody.Find("tr").Each(func(i int, tr *goquery.Selection) {
			if tr.Find("a.comp-type-wca").Length() == 0 {
				return
			}
			cc, err2 := parseCubingCompetition(tr)
			if err2 != nil {
				return
			}
			out = append(out, cc)
		})
	})
	return out, count, nil
}

type UpdateCubingChinaComps struct {
	DB  *gorm.DB
	one sync.Once
}

func (u *UpdateCubingChinaComps) Name() string { return "UpdateCubingChinaComps" }
func (u *UpdateCubingChinaComps) Run() error {
	// 第 1 页：获取数据 + total
	firstPageComps, total, err := getCompetitionWithPage(1)
	if err != nil {
		return fmt.Errorf("failed to fetch page 1: %w", err)
	}
	if total <= 0 {
		return nil
	}
	perPage := 100
	totalPages := (total + perPage - 1) / perPage

	allComps := make([]*CubingCompetition, 0, total)
	allComps = append(allComps, firstPageComps...)
	for page := 2; page <= totalPages; page++ {
		comps, _, err := getCompetitionWithPage(page)
		if err != nil {
			return fmt.Errorf("failed to fetch page %d: %w", page, err)
		}
		if len(comps) == 0 {
			break
		}
		allComps = append(allComps, comps...)
	}

	if len(allComps) != total {
		return fmt.Errorf("failed to fetch page %d: too many cubing competitions", total)
	}

	if err = system.SetKeyJSONValue(u.DB, UpdateCubingCompetitionKey, allComps, "粗饼比赛列表数据"); err != nil {
		return fmt.Errorf("failed to set system key: %w", err)
	}
	log.Infof("set cubing system key ok %d", total)
	return nil
}
