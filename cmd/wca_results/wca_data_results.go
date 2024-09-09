package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/wca_model/utils"
	xutils "github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/go-tables/table"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

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

var events = []string{
	"333", "222", "444", "555", "666", "777",
	"333oh", "333bf",
	"minx", "pyram", "skewb", "sq1", "clock",
	"444bf", "555bf",
	"333fm",
	//"333mbf",
}

const PersonTableName = "Persons"
const ResultTableName = "Results"
const BestTableName = "ConciseAverageResults"
const AvgTableName = "ConciseSingleResults"
const CacheFile = "cache.map.json"
const DbSrc = "root:my123456@tcp(127.0.0.1:3306)/wca_dev?charset=utf8&parseTime=True&loc=Local"
const wcaIDFile = "./names.txt"

func getWcaIDs() []string {

	f, err := os.ReadFile(wcaIDFile)
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
	//fmt.Println(line)
	return line
}

var DB = func() *gorm.DB {
	Db, err := gorm.Open(
		mysql.New(mysql.Config{DSN: DbSrc}), &gorm.Config{
			//Logger: logger.Discard,
		},
	)
	if err != nil {
		panic(err)
	}
	return Db
}()

//
//func getPersonBestResultsMap(wcaID string) (PersonBestResults, error) {
//	var person Persons
//	if err := DB.Table(PersonTableName).First(&person, "wca_id = ?", wcaID).Error; err != nil {
//		return PersonBestResults{}, err
//	}
//
//	var bestResults []Results
//	var avgResults []Results
//
//	DB.Table(BestTableName).Where("personId = ?", wcaID).Find(&bestResults)
//	DB.Table(AvgTableName).Where("personId = ?", wcaID).Find(&avgResults)
//
//	if len(bestResults) == 0 && len(avgResults) == 0 {
//		return PersonBestResults{}, errors.New("not found")
//	}
//
//	var out = PersonBestResults{
//		PersonName: bestResults[0].PersonName,
//		Best:       make(map[string]Results),
//		Avg:        make(map[string]Results),
//	}
//
//	for _, result := range bestResults {
//		eid := result.EventId
//		if !slices.Contains(events, eid) {
//			continue
//		}
//		result.PersonName = person.Name
//		// 处理best为空时
//		if result.Best <= 0 {
//			continue
//		}
//		if oldBest, ok := out.Best[eid]; !ok || result.Best <= oldBest.Best {
//			out.Best[eid] = result
//			continue
//		}
//	}
//
//	for _, result := range avgResults {
//		eid := result.EventId
//		if !slices.Contains(events, eid) {
//			continue
//		}
//		result.PersonName = person.Name
//		if result.Average <= 0 {
//			continue
//		}
//		if oldAvg, ok := out.Avg[eid]; !ok || result.Average <= oldAvg.Average {
//			out.Avg[eid] = result
//		}
//	}
//
//	return out, nil
//}

func getPersonBestResultsMap(wcaID string) (PersonBestResults, error) {
	var person Persons
	if err := DB.Table(PersonTableName).First(&person, "wca_id = ?", wcaID).Error; err != nil {
		return PersonBestResults{}, err
	}

	var results []Results
	DB.Table(ResultTableName).Where("personId = ?", wcaID).Find(&results)
	if len(results) == 0 {
		return PersonBestResults{}, errors.New("not found results")
	}

	var out = PersonBestResults{
		PersonName: person.Name,
		Best:       make(map[string]Results),
		Avg:        make(map[string]Results),
	}

	for _, result := range results {
		eid := result.EventId
		if !slices.Contains(events, eid) {
			continue
		}
		result.PersonName = person.Name
		// 处理best为空时
		if result.Best <= 0 {
			continue
		}
		if oldBest, ok := out.Best[eid]; !ok || result.Best <= oldBest.Best {
			out.Best[eid] = result
			continue
		}
		if result.Average <= 0 {
			continue
		}
		if oldAvg, ok := out.Avg[eid]; !ok || result.Average <= oldAvg.Average {
			out.Avg[eid] = result
		}
	}

	return out, nil
}

func readCacheBestResultsMap() (map[string]PersonBestResults, error) {
	var out = make(map[string]PersonBestResults)

	f, err := os.ReadFile(CacheFile)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(f, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func getAllPersonBestResultsMap(cache bool) map[string]PersonBestResults {
	if cache {
		out, err := readCacheBestResultsMap()
		if err == nil {
			return out
		}
	}

	ts := time.Now()
	var out = make(map[string]PersonBestResults)
	for _, id := range getWcaIDs() {

		r, err := getPersonBestResultsMap(id)
		if err != nil {
			fmt.Println(err, id)
			continue
		}
		out[r.PersonName] = r
	}
	fmt.Printf("use time %v\n", time.Since(ts))

	if cache {
		data, _ := json.MarshalIndent(out, "", "    ")
		_ = os.WriteFile(CacheFile, data, 0644)
	}
	return out
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
				Best:     utils.ResultsTimeFormat(b.Best, evId),
				BestName: getCnName(b.PersonName),
			},
		)
	}
	for idx, a := range avgs {
		tbs[idx].Avg = utils.ResultsTimeFormat(a.Average, evId)
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

func allEventPrint() {
	datas := getAllPersonBestResultsMap(true)
	num := 10
	for _, eid := range events {
		fmt.Printf("=================== %s 前%d排名 ==================\n", eid, num)
		printWithEventTable(eid, datas, num)
	}
}

func main() {
	allEventPrint()
}
