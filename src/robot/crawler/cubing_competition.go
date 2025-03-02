package crawler

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
	sleepDuration := time.Duration(rand.Intn(2001)) * time.Millisecond
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

func getAllCompetitionUrls() []string {
	var out []string
	for i := startYear; i <= endYear; i++ {
		d := getBaseCompetitionUrls(i)
		out = append(out, d...)
		RandSleep()
		fmt.Println(i)
	}
	return out
}

func getPage(url string) (string, bool, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("[e] %s", err)
		return "", false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return url, true, nil
	}
	return "", false, nil
}

func CheckAllCompetition() (find []string) {
	var oldCp = make(map[string]string)
	var newCp = make(map[string]string)
	for _, url := range getAllCompetitionUrls() {
		val := strings.ReplaceAll(url, competitionBase, "")
		oldCp[val] = url
		newVal := val[:len(val)-4]
		newCp[newVal] = url
	}

	for npKey, _ := range newCp {
		nKey := fmt.Sprintf("%s%d", npKey, nowYear)
		if _, ok := oldCp[nKey]; ok {
			continue
		}
		url, isFind, _ := getPage(fmt.Sprintf("%s%s", competitionBase, nKey))
		RandSleep()
		if isFind {
			find = append(find, url)
			fmt.Printf("find => %s\n", url)
		}
	}
	return find
}
