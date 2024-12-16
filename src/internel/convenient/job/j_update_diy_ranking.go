package job

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	"gorm.io/gorm"
	"net/http"
	"path"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DiyRankingsKey = "diy_rankings"
	urlFormat      = "https://www.worldcubeassociation.org/persons/%s" // 2017XUYO01
)

type (
	EventData struct {
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

	PersonResult struct {
		WCA    string
		Name   string
		Events []EventData
	}

	Persons struct {
		WcaId string `gorm:"column:wca_id"`
		Name  string `gorm:"column:name"`
	}

	Results struct {
		EventId    string  `json:"eventId"`
		Best       float64 `json:"best"`
		BestStr    string  `json:"bestStr"`
		Average    float64 `json:"average"`
		AverageStr string  `json:"averageStr"`
		PersonName string  `json:"personName"`
		PersonId   string  `json:"personId"`
	}

	PersonBestResults struct {
		PersonName string             `json:"PersonName"`
		Best       map[string]Results `json:"Best"`
		Avg        map[string]Results `json:"Avg"`
	}
	WcaResult struct {
		BestRank        int    `json:"BestRank"`
		BestStr         string `json:"BestStr"`
		BestPersonName  string `json:"BestPersonName"`
		BestPersonWCAID string `json:"BestPersonWCAID"`
		AvgRank         int    `json:"AvgRank"`
		AvgStr          string `json:"AvgStr"`
		AvgPersonName   string `json:"AvgPersonName"`
		AvgPersonWCAID  string `json:"AvgPersonWCAID"`
	}
)

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

type UpdateDiyRankings struct {
	DB *gorm.DB
}

func (u *UpdateDiyRankings) getWCAResults(wcaID string) (*PersonResult, error) {
	// 请求网页
	resp, err := http.Get(fmt.Sprintf(urlFormat, wcaID))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

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

func (u *UpdateDiyRankings) getCnName(input string) string {
	re := regexp.MustCompile(`\((.*?)\)`)
	match := re.FindStringSubmatch(input)

	if len(match) > 1 {
		return match[1]
	}
	return input
}

func (u *UpdateDiyRankings) parserTimeToSeconds(t string) float64 {
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

func (u *UpdateDiyRankings) getAllPersonBestResultsMap(allP []*PersonResult) map[string]PersonBestResults {
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
				Best:       u.parserTimeToSeconds(val.Single),
				BestStr:    val.Single,
				PersonName: p.Name,
				PersonId:   p.WCA,
			}
			if val.Average != "" {
				pbr.Avg[val.Event] = Results{
					EventId:    val.Event,
					Average:    u.parserTimeToSeconds(val.Average),
					AverageStr: val.Average,
					PersonName: p.Name,
					PersonId:   p.WCA,
				}
			}
		}
		out[p.Name] = pbr
	}
	return out
}

func (u *UpdateDiyRankings) getAllResult(WcaIDs []string) map[string]PersonBestResults {
	var allP []*PersonResult
	WcaIDs = utils.RemoveRepeatedElement(WcaIDs)
	resultsCh := make(chan *PersonResult, len(WcaIDs))
	errCh := make(chan error, len(WcaIDs))
	var wg sync.WaitGroup

	semaphore := make(chan struct{}, 8)

	for _, wcaId := range WcaIDs {
		wg.Add(1)
		go func(id string) {
			semaphore <- struct{}{}
			defer wg.Done()
			res, err := u.getWCAResults(id)
			if err != nil {
				errCh <- err
				return
			}
			resultsCh <- res
			<-semaphore
		}(wcaId)
	}
	wg.Wait()
	close(resultsCh)
	close(errCh)
	for res := range resultsCh {
		allP = append(allP, res)
	}
	for err := range errCh {
		fmt.Println("Error fetching result:", err)
	}
	return u.getAllPersonBestResultsMap(allP)
}

func (u *UpdateDiyRankings) getSorAllResults(wcaIDs []string) map[string][]WcaResult {
	var out = make(map[string][]WcaResult)
	data := u.getAllResult(wcaIDs)

	for _, eid := range eventsList {
		var bests []Results
		var avgs []Results

		for _, r := range data {
			if b, ok := r.Best[eid]; ok {
				bests = append(bests, b)
			}
			if a, ok := r.Avg[eid]; ok {
				avgs = append(avgs, a)
			}
		}
		sort.Slice(bests, func(i, j int) bool { return bests[i].Best < bests[j].Best })
		sort.Slice(avgs, func(i, j int) bool { return avgs[i].Average < avgs[j].Average })

		var wrs []WcaResult
		for idx, b := range bests {
			var index = idx + 1
			if idx >= 1 && wrs[idx-1].BestStr == b.BestStr {
				index = wrs[idx-1].BestRank
			}
			wrs = append(
				wrs, WcaResult{
					BestRank:        index,
					BestStr:         b.BestStr,
					BestPersonName:  b.PersonName,
					BestPersonWCAID: b.PersonId,
				},
			)
		}

		for idx, a := range avgs {
			var index = idx + 1

			if idx >= 1 && wrs[idx-1].AvgStr == a.AverageStr {
				index = wrs[idx-1].AvgRank
			}
			wrs[idx].AvgRank = index
			wrs[idx].AvgStr = a.AverageStr
			wrs[idx].AvgPersonName = a.PersonName
			wrs[idx].AvgPersonWCAID = a.PersonId
		}
		out[eid] = wrs
	}
	return out
}

func (u *UpdateDiyRankings) Name() string { return "UpdateDiyRankings" }

func (u *UpdateDiyRankings) Run() error {
	var keys []string
	if err := system.GetKeyJSONValue(u.DB, DiyRankingsKey, &keys); err != nil {
		return err
	}
	for _, key := range keys {
		var wcaKeys []string
		if err := system.GetKeyJSONValue(u.DB, key, &wcaKeys); err != nil {
			continue
		}

		dataKey := path.Join(DiyRankingsKey, key, "data")
		var kv system.KeyValue
		if err := system.GetKeyJSONValue(u.DB, dataKey, &kv); err == nil {
			if time.Since(kv.UpdatedAt) > time.Minute*15 {
				continue
			}
		}
		data := u.getSorAllResults(wcaKeys)
		_ = system.SetKeyJSONValue(u.DB, dataKey, data, "")
		fmt.Printf("[UpdateDiyRankings] 更新数据 %s\n", key)
	}
	return nil
}
