package models

import (
	"fmt"

	"github.com/guojia99/go-tables/table"
)

type WCAResults struct {
	Id            int    `json:"id"`
	Best          int    `json:"best"`
	Average       int    `json:"average"`
	Name          string `json:"name"`
	CompetitionId string `json:"competition_id"`
	EventId       string `json:"event_id"`
	WcaId         string `json:"wca_id"`
	Attempts      []int  `json:"attempts"`
	BestIndex     int    `json:"best_index"`
	WorstIndex    int    `json:"worst_index"`
}

type (
	Results struct {
		EventId       string `json:"eventId"`
		Best          int    `json:"best"`
		BestStr       string `json:"bestStr"`
		Average       int    `json:"average"`
		AverageStr    string `json:"averageStr"`
		PersonName    string `json:"personName"`
		PersonId      string `json:"personId"`
		WorldRank     int    `json:"world_rank"`
		ContinentRank int    `json:"continent_rank"`
		CountryRank   int    `json:"country_rank"`
	}

	MedalCount struct {
		Gold   int `json:"gold"`
		Silver int `json:"silver"`
		Bronze int `json:"bronze"`
		Total  int `json:"total"`
	}

	RecordCount struct {
		National    int `json:"national"`
		Continental int `json:"continental"`
		World       int `json:"world"`
		Total       int `json:"total"`
	}

	PersonBestResults struct {
		DBVersion string `json:"db_version"` // 数据库查询版本

		PersonName       string             `json:"PersonName"`
		WCAID            string             `json:"wcaId"`
		Best             map[string]Results `json:"Best"`
		Avg              map[string]Results `json:"Avg"`
		CompetitionCount int                `json:"competition_count"`
		MedalCount       MedalCount         `json:"MedalCount"`
		RecordCount      RecordCount        `json:"RecordCount"`
	}
)

func (r *Results) BestString() string {
	if r.WorldRank <= 20 {
		return fmt.Sprintf("%s (WR%d)", r.BestStr, r.WorldRank)
	}
	if r.ContinentRank <= 20 {
		return fmt.Sprintf("%s (CR%d)", r.BestStr, r.ContinentRank)
	}
	if r.CountryRank <= 50 {
		return fmt.Sprintf("%s (NR%d)", r.BestStr, r.CountryRank)
	}
	return r.BestStr
}

func (r *Results) AvgString() string {

	if r.WorldRank <= 20 {
		return fmt.Sprintf("%s (WR%d)", r.AverageStr, r.WorldRank)
	}
	if r.ContinentRank <= 20 {
		return fmt.Sprintf("%s (CR%d)", r.AverageStr, r.ContinentRank)
	}
	if r.CountryRank <= 100 {
		return fmt.Sprintf("%s (NR%d)", r.AverageStr, r.CountryRank)
	}
	return r.AverageStr
}

func (m *MedalCount) String() string {
	if m.Total == 0 {
		return ""
	}

	out := "\n"
	if m.Gold != 0 {
		out += fmt.Sprintf("🥇 冠军数: %d\n", m.Gold)
	}
	if m.Silver != 0 {
		out += fmt.Sprintf("🥈 亚军数: %d\n", m.Silver)
	}
	if m.Bronze != 0 {
		out += fmt.Sprintf("🥉 季军数: %d\n", m.Bronze)
	}
	return out
}

func (r *RecordCount) String() string {
	if r.Total == 0 {
		return ""
	}
	out := "\n"
	if r.World != 0 {
		out += fmt.Sprintf("🌍 世界记录: %d\n", r.World)
	}
	if r.Continental != 0 {
		out += fmt.Sprintf("🌏 洲际记录: %d\n", r.Continental)
	}
	if r.National != 0 {
		out += fmt.Sprintf("🏳️ 国家记录: %d\n", r.National)
	}
	return out
}

var WcaEventsList = []string{
	"333",
	"222",
	"444",
	"555",
	"666",
	"777",
	"333bf",
	"333fm",
	"333oh",
	"clock",
	"minx",
	"pyram",
	"skewb",
	"sq1",
	"444bf",
	"555bf",
	"333mbf",
}

var WcaEventsCnMap = map[string]string{
	"333":    "三阶",
	"222":    "二阶",
	"444":    "四阶",
	"555":    "五阶",
	"666":    "六阶",
	"777":    "七阶",
	"333bf":  "三盲",
	"333fm":  "最少步",
	"333oh":  "单手",
	"clock":  "魔表",
	"minx":   "五魔",
	"pyram":  "金字塔",
	"skewb":  "斜转",
	"sq1":    "SQ-1",
	"444bf":  "四盲",
	"555bf":  "五盲",
	"333mbf": "多盲",
}

type personBestResultsTable struct {
	Ev   string `json:"ev"`
	Best string `json:"best"`
	LL   string `json:"ll"`
	Avg  string `json:"avg"`
}

func (s *PersonBestResults) String() string {
	out := "\n"
	out += s.PersonName + "\n"
	out += s.WCAID + "\n"
	out += fmt.Sprintf("参赛次数: %d\n", s.CompetitionCount)
	// 成绩

	mbfOut := ""

	var fullEven = true
	var tbs []personBestResultsTable
	for _, ev := range WcaEventsList {
		b, hasB := s.Best[ev]
		if !hasB {
			fullEven = false
			continue
		}

		if ev == "333mbf" {
			mbfOut += fmt.Sprintf("   %s  %s\n", WcaEventsCnMap[ev], b.BestString())
			continue
		}
		tb := personBestResultsTable{
			Ev:   WcaEventsCnMap[ev],
			Best: b.BestString(),
		}
		a, hasA := s.Avg[ev]
		if hasA {
			tb.LL = " || "
			tb.Avg = a.AvgString()
		} else {
			fullEven = false
		}
		tbs = append(tbs, tb)
	}

	if fullEven {
		out += fmt.Sprintf("⭐⭐⭐大满贯选手⭐⭐⭐\n")
	}

	out += "================\n"
	tb, _ := table.SimpleTable(tbs, &table.Option{
		ExpendID: false,
		Align:    table.AlignCenter,
		Contour:  table.EmptyContour,
	})

	tb.Headers = make(table.RowCell, 0)
	out += tb.String()
	out += mbfOut

	if m := s.MedalCount.String(); m != "" {
		out += "================"
		out += m
	}

	if r := s.RecordCount.String(); r != "" {
		out += "================"
		out += r
	}
	return out
}
