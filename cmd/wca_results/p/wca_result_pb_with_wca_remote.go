package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/guojia99/go-tables/table"
)

func RemoveRepeatedElement[S ~[]E, E comparable](s S) S {
	result := make([]E, 0)
	m := make(map[E]bool)
	for _, v := range s {
		if _, ok := m[v]; !ok {
			result = append(result, v)
			m[v] = true
		}
	}
	return result
}

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
	line = RemoveRepeatedElement(line)
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

type Persons struct {
	WcaId string `gorm:"column:wca_id"`
	Name  string `gorm:"column:name"`
}

type Results struct {
	EventId    string `json:"eventId" gorm:"column:eventId"`
	Best       int    `json:"best" gorm:"column:best"`
	Average    int    `json:"average" gorm:"column:average"`
	PersonName string `json:"personName" gorm:"column:personName"`
	PersonId   string `json:"personId" gorm:"column:personId"`
}

type PersonBestResults struct {
	PersonName string             `json:"PersonName"`
	Best       map[string]Results `json:"Best"`
	Avg        map[string]Results `json:"Avg"`
}

type TableV struct {
	Num      int    `table:"排名"`
	Best     string `table:""`
	BestName string `table:"单次"`
	AvgName  string `table:"平均"`
	Avg      string `table:""`
}

func getCnName(input string) string {
	re := regexp.MustCompile(`\((.*?)\)`)
	match := re.FindStringSubmatch(input)

	if len(match) > 1 {
		return match[1]
	}
	return input
}
func SecondTimeFormat(seconds float64, mbf bool) string {
	intSeconds := int64(seconds)
	decimalSeconds := int64(seconds*100) % 100
	duration := time.Duration(intSeconds) * time.Second

	hours := int64(duration.Hours())
	minutes := int64(duration.Minutes()) % 60
	secondsInt := int64(duration.Seconds()) % 60

	mmSecondsStr := fmt.Sprintf(".%02d", decimalSeconds)
	if decimalSeconds == 0 && (duration >= time.Hour || mbf) {
		mmSecondsStr = ""
	}
	//fmt.Println(fmt.Sprintf("%d:%02d:%02d%s", hours, minutes, secondsInt, mmSecondsStr))
	//return strings.TrimLeft(fmt.Sprintf("%d:%02d:%02d%s", hours, minutes, secondsInt, mmSecondsStr), "0:")
	if duration < time.Minute {
		return fmt.Sprintf("%d%s", secondsInt, mmSecondsStr)
	}
	if duration < time.Hour {
		return fmt.Sprintf("%d:%02d%s", minutes, secondsInt, mmSecondsStr)
	}
	return fmt.Sprintf("%d:%02d:%02d%s", hours, minutes, secondsInt, mmSecondsStr)
}

func ResultsTimeFormat(in int, event string) string {
	switch in {
	case -1:
		return "DNF"
	case -2:
		return "DNS"
		// todo other particular result
	default:
	}

	switch event {
	case "333fm":
		if in > 1000 {
			return fmt.Sprintf("%.2f", float64(in)/100.0)
		}
		return fmt.Sprintf("%d", in)
	case "333mbf":
		// https://www.worldcubeassociation.org/export/results
		//difference    = 99 - DD
		//timeInSeconds = TTTTT (99999 means unknown)
		//missed        = MM
		//solved        = difference + missed
		//attempted     = solved + missed
		strIn := strconv.Itoa(in)
		diff, _ := strconv.Atoi(strIn[:2])
		miss, _ := strconv.Atoi(strIn[len(strIn)-2:])
		seconds, _ := strconv.Atoi(strIn[3 : len(strIn)-2])
		if seconds == 99999 {
			return "unknown"
		}
		formattedTime := SecondTimeFormat(float64(seconds), true)
		solved := 99 - diff + miss
		attempted := solved + miss
		return fmt.Sprintf("%d/%d %s", solved, attempted, formattedTime)
	}
	return SecondTimeFormat(float64(in)/100.0, false)
}

func printWithEventTable(evId string, datas map[string]PersonBestResults, max int) {
	var bests []Results
	var avgs []Results
	for _, d := range datas {
		if b, ok := d.Best[evId]; ok {
			bests = append(bests, b)
		}
		if a, ok := d.Avg[evId]; ok {
			avgs = append(avgs, a)
		}
	}

	sort.Slice(bests, func(i, j int) bool { return bests[i].Best < bests[j].Best })
	sort.Slice(avgs, func(i, j int) bool { return avgs[i].Average < avgs[j].Average })

	var tbs []TableV
	for idx, b := range bests {
		tbs = append(
			tbs, TableV{
				Num:      idx + 1,
				Best:     ResultsTimeFormat(b.Best, evId),
				BestName: getCnName(b.PersonName),
			},
		)
	}
	for idx, a := range avgs {
		tbs[idx].Avg = ResultsTimeFormat(a.Average, evId)
		tbs[idx].AvgName = getCnName(a.PersonName)
	}

	if len(tbs) > max {
		tbs = tbs[:max]
	}

	tb, _ := table.SimpleTable(
		tbs, &table.Option{
			//ExpendID: true,
			Align:   table.AlignCenter,
			Contour: table.DefaultContour,
		},
	)
	fmt.Println(tb)
}

func ParserTimeToSeconds(t string) float64 {
	// 解析纯秒数格式
	if regexp.MustCompile(`^\d+(\.\d+)?$`).MatchString(t) {
		seconds, _ := strconv.ParseFloat(t, 64)
		return seconds
	}

	// 解析分+秒格式
	if regexp.MustCompile(`^\d{1,3}:\d{1,3}(\.\d+)?$`).MatchString(t) {
		parts := strings.Split(t, ":")
		minutes, _ := strconv.ParseFloat(parts[0], 64)
		seconds, _ := strconv.ParseFloat(parts[1], 64)
		return minutes*60 + seconds
	}

	// 解析时+分+秒格式
	if regexp.MustCompile(`^\d{1,3}:\d{1,3}:\d{1,3}(\.\d+)?$`).MatchString(t) {
		parts := strings.Split(t, ":")
		hours, _ := strconv.ParseFloat(parts[0], 64)
		minutes, _ := strconv.ParseFloat(parts[1], 64)
		seconds, _ := strconv.ParseFloat(parts[2], 64)
		return hours*3600 + minutes*60 + seconds
	}

	return -1
}

func getAllPersonBestResultsMap() map[string]PersonBestResults {
	var allP []*PersonResult
	f, _ := os.ReadFile("./cache.json")
	_ = json.Unmarshal(f, &allP)
	fmt.Println(len(allP))

	var out = make(map[string]PersonBestResults)

	for _, p := range allP {
		pbr := PersonBestResults{
			PersonName: p.Name,
			Best:       make(map[string]Results),
			Avg:        make(map[string]Results),
		}

		for _, val := range p.Events {
			if !slices.Contains(eventsList, val.Event) {
				continue
			}
			pbr.Best[val.Event] = Results{
				EventId:    val.Event,
				Best:       int(ParserTimeToSeconds(val.Single) * 100),
				PersonName: p.Name,
				PersonId:   p.WCA,
			}
			if val.Average != "" {
				pbr.Avg[val.Event] = Results{
					EventId:    val.Event,
					Average:    int(ParserTimeToSeconds(val.Average) * 100),
					PersonName: p.Name,
					PersonId:   p.WCA,
				}
			}
		}
		out[p.Name] = pbr
	}
	return out
}

func allEventPrint() {
	datas := getAllPersonBestResultsMap()

	num := 10
	for _, eid := range eventsList {
		fmt.Printf("=================== %s 前%d排名 ==================\n", eid, num)
		printWithEventTable(eid, datas, num)
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
	allEventPrint()
}
