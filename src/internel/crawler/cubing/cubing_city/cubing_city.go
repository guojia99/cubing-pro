package cubing_city

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

const (
	competitionUrl      = "https://cubing.com/competition?year=%d&type=WCA&province=&event="
	competitionBase     = "https://cubing.com/competition/"
	competitionLiveBase = "https://cubing.com/live/"
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

func getAllCompetitionUrls(startYear, endYear int) []string {
	var out []string
	for i := startYear; i <= endYear; i++ {
		d := getBaseCompetitionUrls(i)
		out = append(out, d...)
	}
	return out
}

var notCityKey = []string{
	"China", "Chinas",
	"PKU", "Hong",
	"Alpha", "Cyclops",
	"Floating", "One",
	"WCA", "Tibet",
	"2015", "Cube", "Parity",
	"Please", "GDSY", "XJTU", "Special",
	"Coconut", "HDU", "TKK", "TJU", "ZHBIT",
	"WHU", "FMC", "Big", "Cross", "AHAU",
}

// GetCubingCityListAndOldKey 获取粗饼城市列表和旧的所有历史的key(已经在列表上的不处理)
func GetCubingCityListAndOldKey(startYear, endYear int) (cityList []string, oldKey []string) {
	list := getAllCompetitionUrls(startYear, endYear)
	//out, _ := json.Marshal(list)
	//_ = os.WriteFile("test.json", out, 0644)
	//var list []string
	//out, _ := os.ReadFile("test.json")
	//_ = json.Unmarshal(out, &list)

	for _, nowComp := range list {
		// 旧的Key
		k := utils.ReplaceAll(nowComp, "", competitionBase, competitionLiveBase)
		oldKey = append(oldKey, k)

		if len(k) == 0 {
			continue
		}

		first := strings.Split(k, "-")[0]
		if utils.ContainsString(first, notCityKey...) {
			continue
		}
		cityList = append(cityList, first)
	}
	cityList = utils.RemoveDuplicates(cityList)
	return
}
