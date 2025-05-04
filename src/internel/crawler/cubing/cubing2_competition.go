package cubing

import (
	"bytes"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/guojia99/cubing-pro/src/internel/crawler/cubing/cubing_city"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type DCubingCompetition struct {
	startYear int
	endYear   int

	currYear int
	nexYear  int

	cubingCity []string
	oldKey     map[string]bool // 原本就有的比赛Key
}

func NewDCubingCompetition() *DCubingCompetition {
	c := &DCubingCompetition{
		startYear: 2015,
		endYear:   2025,
	}
	c.updateYear()
	c.updateCityList()
	return c
}

func (c *DCubingCompetition) updateYear() {
	currentYear := time.Now().Year()
	currentMonth := int(time.Now().Month())
	c.currYear = currentYear
	c.nexYear = currentYear + 1
	c.endYear = currentYear
	if currentMonth >= 9 {
		c.endYear += 1
	}
}

func (c *DCubingCompetition) updateCityList() {
	c.oldKey = make(map[string]bool)
	c.cubingCity = make([]string, 0)

	// 粗饼城市
	cubingCity, oldKeys := cubing_city.GetCubingCityListAndOldKey(c.startYear, c.endYear)
	for _, o := range oldKeys {
		c.oldKey[o] = true
	}
	c.cubingCity = append(c.cubingCity, cubingCity...)

	// one城市
	c.cubingCity = append(c.cubingCity, cubing_city.GetOneCityList()...)

	// 其他补充的city
	c.cubingCity = append(c.cubingCity, cubing_city.OtherCitys()...)

	c.cubingCity = utils.RemoveDuplicates(c.cubingCity)
}

// CompNameWithMouth 第一个为 今年， 第二个为明年
var CompNameWithMouth = map[string][2][]int{
	"Open":        {{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, {10, 11, 12}},
	"Spring":      {{1, 2, 3}, {10, 11, 12}},
	"Spring-Open": {{1, 2, 3}, {10, 11, 12}},
	"Summer":      {{2, 3, 4, 5, 6}, {}},
	"Summer-Open": {{2, 3, 4, 5, 6}, {}},
	"Autumn":      {{6, 7, 8, 9, 10}, {}},
	"Autumn-Open": {{6, 7, 8, 9, 10}, {}},
	"Winter":      {{1, 2}, {9, 10, 11, 12}},
	"Winter-Open": {{1, 2}, {9, 10, 11, 12}},
	"Newcomers":   {{1, 2}, {12}},
	"New-Year":    {{1, 2}, {11, 12}},
}

func (c *DCubingCompetition) newKeyWithCity() []string {
	currentMonth := int(time.Now().Month())
	var outKeys []string

	for key, ms := range CompNameWithMouth {
		curYearMouth := ms[0]
		nextYearMouth := ms[1]

		for _, city := range c.cubingCity {
			if slices.Contains(curYearMouth, currentMonth) {
				outKeys = append(outKeys, fmt.Sprintf("%s-%s-%d", city, key, c.currYear))
			}
			if slices.Contains(nextYearMouth, currentMonth) {
				outKeys = append(outKeys, fmt.Sprintf("%s-%s-%d", city, key, c.nexYear))
			}
		}
	}

	for _, key := range cubing_city.OtherKeys() {
		outKeys = append(outKeys, fmt.Sprintf("%s-%d", key, c.currYear))
		outKeys = append(outKeys, fmt.Sprintf("%s-%d", key, c.nexYear))
	}

	outKeys = utils.RemoveDuplicates(outKeys)
	return outKeys
}

type TCubingCompetition struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Url    string
	Date   string `json:"date"`
	Events string `json:"events"`
}

func (c *DCubingCompetition) getPage(id, url string) (TCubingCompetition, bool, error) {
	resp, err := utils.HTTPRequestFull("GET", url, nil, map[string]interface{}{
		"Cache-Control":             "max-age=0, private, must-revalidate",
		"Content-Encoding":          "gzip, deflate, br, zstd",
		"Accept-Language":           "zh-CN,zh-HK;q=0.9,zh;q=0.8,zh-TW;q=0.7,en;q=0.6",
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"User-Agent":                "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36",
		"Content-Type":              "text/html; charset=UTF-8",
		"Date":                      "Sun, 04 May 2025 07:12:22 GMT",
		"Eagleid":                   "7ce1a79817463427425055368e",
		"Server":                    "Tengine",
		"Strict-Transport-Security": "max-age=5184000",
		"Timing-Allow-Origin":       "*",
		"Vary":                      "Accept-Encoding",
		"Via":                       "ens-cache33.l2hk12[45,0], cache51.l2so158-1[49,0], kunlun4.cn2466[69,0]",
	}, nil)
	if err != nil {
		return TCubingCompetition{}, false, err
	}

	if resp.StatusCode != 200 {
		return TCubingCompetition{}, false, fmt.Errorf("[e] %d", resp.StatusCode)
	}

	//fmt.Println(string(resp))
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
	if err != nil {
		return TCubingCompetition{}, false, err
	}

	var out = TCubingCompetition{
		ID:  id,
		Url: url,
	}
	doc.Find("h1.heading-title").Each(func(i int, s *goquery.Selection) {
		out.Name = s.Text()
	})
	doc.Find("dt#events").NextFiltered("dd").Each(func(i int, s *goquery.Selection) {
		out.Events = s.Text()
	})
	doc.Find("dt").Each(func(i int, s *goquery.Selection) {
		if strings.TrimSpace(s.Text()) == "日期" {
			dd := s.NextFiltered("dd") // 获取紧随其后的 <dd>
			out.Date = strings.TrimSpace(dd.Text())
		}
	})

	return out, true, nil
}

const competitionBase = "https://cubing.com/competition/"

func (c *DCubingCompetition) GetNewCompetitions() []TCubingCompetition {
	baseKeys := c.newKeyWithCity()

	log.Printf("===> 尝试获取新比赛%d条\n", len(baseKeys))

	var (
		find []TCubingCompetition
		wg   sync.WaitGroup
		mu   sync.Mutex
		ch   = make(chan TCubingCompetition, len(baseKeys)) // 使用缓冲通道存储结果
	)

	idx := 0
	for _, nKey := range baseKeys {
		idx++
		if _, ok := c.oldKey[nKey]; ok {
			continue
		}

		wg.Add(1)
		go func(nKey string, idx int) {
			defer wg.Done()
			pUrl := fmt.Sprintf("%s%s", competitionBase, nKey)
			url, isFind, _ := c.getPage(nKey, pUrl)
			if isFind {
				ch <- url
				log.Printf("=========== find = %s ==> %s\n", nKey, url)
			}
			time.Sleep(time.Millisecond * 100)
			log.Printf("[%d]check => %s | %s\n", idx, pUrl, nKey)
		}(nKey, idx)
	}

	wg.Wait()
	close(ch) // 关闭通道，确保所有 Goroutine 结束

	// 收集结果
	for url := range ch {
		mu.Lock()
		find = append(find, url)
		mu.Unlock()
	}

	return find
}
