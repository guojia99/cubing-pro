package cubing

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func RandSleep() {
	sleepDuration := time.Duration(rand.Intn(2000)) * time.Millisecond
	if sleepDuration <= time.Microsecond*300 {
		sleepDuration = time.Microsecond * 300
	}
	time.Sleep(sleepDuration)
}

const (
	startYear       = 2010
	endYear         = 2025 // todo
	nowYear         = 2025
	competitionUrl  = "https://cubing.com/competition?year=%d&type=WCA&province=&event="
	competitionBase = "https://cubing.com/competition/"
)

// 获取目前已有的比赛状态
func getBaseCompetitionUrls(year int) []string {
	var out []string
	resp, err := http.Get(fmt.Sprintf(competitionUrl, year))
	if err != nil {
		log.Printf("[e] %s", err)
		return out
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return out
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return out
	}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if !exists {
			return
		}
		if !strings.HasPrefix(link, competitionBase) {
			return
		}
		out = append(out, link)
	})
	return out
}

const cacheFile = "competition_urls.json"

func getAllCompetitionUrls() []string {
	var out []string
	for i := startYear; i <= endYear; i++ {
		d := getBaseCompetitionUrls(i)
		out = append(out, d...)
		RandSleep()
	}
	return out
}

type TCubingCompetition struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Url    string
	Date   string `json:"date"`
	Events string `json:"events"`
}

func getPage(id, url string) (TCubingCompetition, bool, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("[e] %s", err)
		return TCubingCompetition{}, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return TCubingCompetition{}, false, fmt.Errorf("[e] %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
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

var cpNames = []string{
	// 公开赛
	"Open",
	"Spring",
	"Spring-Open",
	"Summer",
	"Summer-Open",
	//"Winter",
	//"Winter-Open",
	//"Autumn",
	//"Autumn-Open",

	// 专项赛
	"Big-Cubes",
	"Big-Cubes-Spring",
	"Big-Cubes-Summer",
	"Special",
	//"Quiet-Day", // 安静赛
}

var cpPrefix = []string{
	"Please-Be-Quiet", // 安静赛
}

var probablyOtherCitys = []string{
	"Shijiazhuang", "Taiyuan", "Datong", "Changchun", "Nanjing", "Ningbo",
	"Fuzhou", "Jinan", "Qindao", "Nanning", "Haikou", "Sanya", "Kunming", "Yinchuan", "Weifang",
	"LinYyi", "Yantai", "Wuxi", "Changzhou", "Nantong", "Xiamen", "Quanzhou", "Wenzhou", "Jinhua", "Shaoxing",
	"Baoding", "Zhuhai", "Zhongshan", "Lanzhou",
}

// getAllProbablyUrl 获取可能的组合
func getAllProbablyUrl() (oldCp, newCp map[string]string) {
	oldCp = make(map[string]string)
	newCp = make(map[string]string)

	nowCompUrls := getAllCompetitionUrls()

	// 重新拼接
	for _, url := range nowCompUrls {
		val := strings.ReplaceAll(url, competitionBase, "")
		oldCp[val] = url
		newVal := val[:len(val)-4]
		newCp[newVal] = url
	}

	// 获取所有城市
	for _, url := range nowCompUrls {
		val := strings.ReplaceAll(url, competitionBase, "")
		sp := strings.Split(val, "-")
		// 后缀
		for _, v := range cpNames {
			newVal := fmt.Sprintf("%s-%s-", sp[0], v)
			newCp[newVal] = url
		}
		// 前缀
		for _, v := range cpPrefix {
			if v == sp[0] {
				continue
			}
			newVal := fmt.Sprintf("%s-%s-", v, sp[0])
			newCp[newVal] = url
		}
	}

	// 补充城市
	for _, city := range probablyOtherCitys {
		for _, v := range cpNames {
			newVal := fmt.Sprintf("%s-%s-", city, v)
			newCp[newVal] = city
		}
	}

	return
}

func CheckAllCubingCompetition() (find []TCubingCompetition) {
	var oldCp, newCp = getAllProbablyUrl()
	idx := 0
	for npKey, _ := range newCp {
		idx += 1
		nKey := fmt.Sprintf("%s%d", npKey, nowYear)
		if _, ok := oldCp[nKey]; ok {
			continue
		}
		url, isFind, _ := getPage(nKey, fmt.Sprintf("%s%s", competitionBase, nKey))
		RandSleep()
		time.Sleep(time.Millisecond * 400)
		if isFind {
			find = append(find, url)
			fmt.Printf("=========== find => %s\n", url)
		}
		fmt.Printf("[%d]check => %s\n", idx, nKey)
	}
	return find
}
