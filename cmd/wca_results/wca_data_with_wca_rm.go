package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	xutils "github.com/guojia99/cubing-pro/src/internel/utils"
)

func GetWcaIDs() []string {
	f, err := os.ReadFile("./names.txt")
	if err != nil {
		panic(err)
	}
	data := string(f)

	line := strings.Split(data, "\n")
	for i := range line {
		newLine := strings.Split(line[i], ".")[1]
		line[i] = strings.ReplaceAll(newLine, " ", "")
	}
	line = xutils.RemoveRepeatedElement(line)
	return line
}

type EventData struct {
	Event          string `json:"event"`
	CountryRank    string `json:"country_rank"`
	ContinentRank  string `json:"continent_rank"`
	WorldRank      string `json:"world_rank"`
	Single         string `json:"single"`
	Average        string `json:"average"`
	WorldRank2     string `json:"world_rank_2"`
	ContinentRank2 string `json:"continent_rank_2"`
	CountryRank2   string `json:"country_rank_2"`
}

type PersonResult struct {
	WCA    string
	Name   string
	Events []EventData
}

var eventsList = []string{
	"3x3x3Cube",
	"2x2x2Cube",
	"4x4x4Cube",
	"5x5x5Cube",
	"6x6x6Cube",
	"7x7x7Cube",
	"3x3x3Blindfolded",
	"3x3x3FewestMoves",
	"3x3x3One-Handed",
	"Clock",
	"Megaminx",
	"Pyraminx",
	"Skewb",
	"Square-1",
	"4x4x4Blindfolded",
	"5x5x5Blindfolded",
	//3x3x3Multi-Blind
}

const urlFormat = "https://www.worldcubeassociation.org/persons/%s" // 2017XUYO01

func getWCAResults(wcaID string) (*PersonResult, error) {
	// 请求网页
	resp, err := http.Get(fmt.Sprintf(urlFormat, wcaID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error: status code %d", resp.StatusCode)
	}

	// 解析网页
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	var events []EventData
	// 查找并提取所需的数据 (例如：提取所有链接的文本和地址)
	doc.Find(".personal-records").Each(
		func(_ int, item *goquery.Selection) {
			item.Find(".table-responsive").Each(
				func(_ int, table *goquery.Selection) {
					table.Find("table.table-striped tbody tr").Each(
						func(index int, row *goquery.Selection) {
							// 提取每一列的数据
							event := row.Find("td.event").Text()
							countryRank := row.Find("td.country-rank").First().Text()
							continentRank := row.Find("td.continent-rank").First().Text()
							worldRank := row.Find("td.world-rank").First().Text()
							single := row.Find("td.single").Text()
							average := row.Find("td.average").Text()
							worldRank2 := row.Find("td.world-rank").Last().Text()
							continentRank2 := row.Find("td.continent-rank").Last().Text()
							countryRank2 := row.Find("td.country-rank").Last().Text()

							// 将数据添加到 events 列表中
							events = append(
								events, EventData{
									Event:          strings.ReplaceAll(strings.ReplaceAll(event, "\n", ""), " ", ""),
									CountryRank:    strings.ReplaceAll(strings.ReplaceAll(countryRank, "\n", ""), " ", ""),
									ContinentRank:  strings.ReplaceAll(strings.ReplaceAll(continentRank, "\n", ""), " ", ""),
									WorldRank:      strings.ReplaceAll(strings.ReplaceAll(worldRank, "\n", ""), " ", ""),
									Single:         strings.ReplaceAll(strings.ReplaceAll(single, "\n", ""), " ", ""),
									Average:        strings.ReplaceAll(strings.ReplaceAll(average, "\n", ""), " ", ""),
									WorldRank2:     strings.ReplaceAll(strings.ReplaceAll(worldRank2, "\n", ""), " ", ""),
									ContinentRank2: strings.ReplaceAll(strings.ReplaceAll(continentRank2, "\n", ""), " ", ""),
									CountryRank2:   strings.ReplaceAll(strings.ReplaceAll(countryRank2, "\n", ""), " ", ""),
								},
							)
						},
					)
				},
			)
		},
	)

	var name string
	doc.Find("#person").Each(
		func(i int, selection *goquery.Selection) {
			selection.Find(".text-center").Each(
				func(i int, selection *goquery.Selection) {
					selection.Find("h2").Each(
						func(i int, selection *goquery.Selection) {
							name = selection.Text()
						},
					)
				},
			)
		},
	)

	name = strings.ReplaceAll(strings.ReplaceAll(name, "\n", ""), " ", "")
	//fmt.Printf("%+v\n", events)
	return &PersonResult{
		WCA:    wcaID,
		Name:   name,
		Events: events,
	}, nil
}

func getAllResultsWithFile() {
	var allP []*PersonResult
	f, _ := os.ReadFile("./all_data_%s.json")
	_ = json.Unmarshal(f, &allP)

	for _, event := range eventsList {

	}
}
func main() {
	//var allP []*PersonResult

	//for _, wcaId := range GetWcaIDs() {
	//	fmt.Println(wcaId)
	//	p, err := getWCAResults(wcaId)
	//	time.Sleep(time.Millisecond * 200)
	//	if err != nil {
	//		fmt.Println(wcaId, err)
	//		continue
	//	}
	//	allP = append(allP, p)
	//}
	//data, _ := json.MarshalIndent(allP, "", "    ")
	//os.WriteFile(fmt.Sprintf("./all_data_%s.json", time.Now().UnixNano()), data, 0644)
	getAllResultsWithFile()
}
