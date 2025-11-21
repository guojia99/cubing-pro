package wca_api

import (
	"fmt"

	wca_model "github.com/guojia99/cubing-pro/src/internel/database/model/wca"
	"github.com/guojia99/go-tables/table"
)

type SeniorRank struct {
	Rank        int    `json:"rank"`
	Id          string `json:"id"`
	Best        string `json:"best"`
	Competition int    `json:"competition"`

	// 自定义
	Type string `json:"type"` // single, average
	Age  int    `json:"age"`  // 表示该成绩在什么时候年龄组获取的

	// 排名
	WR int `json:"wr"` // 世界记录排行
	CR int `json:"cr"` // 洲记录排行
	NR int `json:"nr"` // 国家记录排行
}

func (s SeniorRank) String() string {
	if s.WR <= 20 {
		return fmt.Sprintf("%s (WR%d)", s.Best, s.WR)
	}
	if s.CR <= 20 {
		return fmt.Sprintf("%s (CR%d)", s.Best, s.CR)
	}
	if s.NR <= 100 {
		return fmt.Sprintf("%s (NR%d)", s.Best, s.NR)
	}
	return s.Best
}

type SeniorMissing struct {
	World      int            `json:"world"`
	Continents map[string]int `json:"continents"`
	Countries  map[string]int `json:"countries"`
}
type SeniorRanking struct {
	Type    string        `json:"type"` // single, average
	Age     int           `json:"age"`
	Ranks   []SeniorRank  `json:"ranks"`
	Missing SeniorMissing `json:"missing"`
}
type SeniorEvents struct {
	Id       string          `json:"id"`
	Name     string          `json:"name"`
	Format   string          `json:"format"`
	Rankings []SeniorRanking `json:"rankings"`
}

type SeniorPerson struct {
	Id       string   `json:"id"`
	Name     string   `json:"name"`
	Country  string   `json:"country"` // GB 短缩写
	Username string   `json:"username"`
	UserNum  int      `json:"usernum"`
	Age      int      `json:"age"`
	Events   []string `json:"events"`
}

type SeniorCountry struct {
	ID        string `json:"ID"`
	Name      string `json:"name"`
	Continent string `json:"continent"`
}

type SeniorContinent struct {
	ID   string `json:"ID"`
	Name string `json:"name"`
}

type BestSeniorValue struct {
	Single  map[string]SeniorRank `json:"single"`
	Average map[string]SeniorRank `json:"average"`
}

type SeniorPersonValue struct {
	SeniorPerson
	// 自定义增加的内容
	CountryName string `json:"countryName"`
	Continent   string `json:"continent"` // 短缩写

	Single  map[int]map[string]SeniorRank `json:"single"` // 该选手在不同年龄组里获取到的成绩排行
	Average map[int]map[string]SeniorRank `json:"average"`
}

type SeniorsData struct {
	Refreshed  string            `json:"refreshed"`
	Events     []SeniorEvents    `json:"events"`
	Persons    []SeniorPerson    `json:"persons"`
	Countries  []SeniorCountry   `json:"countries"`
	Continents []SeniorContinent `json:"continents"`

	// 自定义
	PersonMap map[string]SeniorPersonValue `json:"personMap"`
}

func getAges(age int) []int {
	var out []int
	for i := age; i >= 40; i -= 10 {
		out = append(out, i)
	}
	return out
}

type personBestResultsTable struct {
	Ev   string `json:"ev"`
	Best string `json:"best"`
	LL   string `json:"ll"`
	Avg  string `json:"avg"`
}

func (p *SeniorPersonValue) String() string {
	out := fmt.Sprintf("%s | %s (%d)\n", p.Id, p.Name, p.Age)

	ages := getAges(p.Age)

	for _, age := range ages {
		if p.Single[age] == nil || len(p.Single[age]) == 0 {
			continue
		}

		singleMap := p.Single[age]
		avgMap := p.Average[age]
		if avgMap == nil {
			avgMap = make(map[string]SeniorRank)
		}
		var tbs []personBestResultsTable
		var mbfOut string

		for _, ev := range wca_model.WcaEventsList {
			cn := wca_model.WcaEventsCnMap[ev]
			single, sOk := singleMap[ev]
			if !sOk {
				continue
			}

			if ev == "333mbf" {
				mbfOut += fmt.Sprintf("   %s  %s\n", cn, single.String())
				continue
			}

			tb := personBestResultsTable{
				Ev:   cn,
				Best: single.String(),
			}

			avg, hasAvg := avgMap[ev]
			if hasAvg {
				tb.LL = " || "
				tb.Avg = avg.String()
			}
			tbs = append(tbs, tb)
		}

		if len(mbfOut) > 0 || len(tbs) > 0 {
			out += fmt.Sprintf("========= %d =========\n", age)
			tb, _ := table.SimpleTable(tbs, &table.Option{
				ExpendID: false,
				Align:    table.AlignCenter,
				Contour:  table.EmptyContour,
			})
			tb.Headers = make(table.RowCell, 0)
			out += tb.String()
			out += mbfOut
		}
	}

	out += "\n 数据来源wca seniors org"
	return out
}
