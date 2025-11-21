package staticx

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

type Fields struct {
	Country string

	A1Comps, A2Rounds, A3Cp         int
	A4WR, A5AsR, A6NR               int
	B1Persons, B2Persons, B3Persons int
	C1WR, C2AsR, C3NR               int
}

func (s *StaticX) GetStatic(country string, year int) string {
	//	format := `这一年有（A1）场WCA赛事在中国大陆举行，一共举行了（A2）个项目轮次，其中诞生了(A3)个赛事冠军。
	//有(B1)名中国选手在今年参加了WCA比赛，其中1522名中国选手是在今年首次参加了WCA赛事。
	//在今年的WCA比赛中，中国选手一共破了世界纪录(WR) (C1)个，亚洲纪录(AsR) (C2)个，国家纪录(NR) (C3)个。`
	format := `这一年有（{{.A1Comps}}）场WCA赛事在{{.Country}}举行，一共举行了（{{.A2Rounds}}）个项目轮次，其中诞生了({{.A3Cp}})个赛事冠军。
在国内赛事中,一共打破了世界纪录(WR) ({{.A4WR}})个，亚洲纪录(AsR) ({{.A5AsR}})个，国家纪录(NR) ({{.A6NR}})个。

有({{.B1Persons}})名{{.Country}}选手今年在{{.Country}}参加了WCA比赛。
有({{.B2Persons}})名{{.Country}}选手在全世界参加了WCA比赛, 其中({{.B3Persons}})名{{.Country}}选手是在今年首次参加了WCA赛事。

在今年的WCA比赛中，{{.Country}}选手一共破了世界纪录(WR) ({{.C1WR}})个，亚洲纪录(AsR) ({{.C2AsR}})个，国家纪录(NR) ({{.C3NR}})个。`

	fs := Fields{
		Country: country,
	}

	var competitions []Competition
	var resultsMap = make(map[string][]Result)
	var allWithCountryResult []Result

	s.db.Where("countryId = ?", country).Where("year = ?", year).Find(&competitions)
	for _, competition := range competitions {
		var compResults []Result
		s.db.Where("competitionId = ?", competition.ID).Find(&compResults)
		resultsMap[competition.ID] = compResults
	}
	s.db.Where("competitionId like ? and personCountryId = ?", fmt.Sprintf("%%%d", year), country).Find(&allWithCountryResult)
	addRWithCountry := func(record *string) {
		if record == nil {
			return
		}
		switch *record {
		case "WR":
			fs.A4WR += 1
		case "AsR":
			fs.A5AsR += 1
		case "NR":
			fs.A6NR += 1
		}
	}

	var roundMap = make(map[string]struct{})  // 轮次 CompId-EventID-PersonType
	var personMap = make(map[string]struct{}) // 人
	var cpMap = make(map[string]struct{})     // 赛事冠军 CompId-EventID-PersonName
	for compId, results := range resultsMap {
		for _, result := range results {
			roundMap[fmt.Sprintf("%s-%s-%s", compId, result.EventID, result.RoundTypeID)] = struct{}{}
			// 赛事冠军， Best不DNF， 且
			if result.Best > 0 && result.Pos == 1 && (result.RoundTypeID == "f" || result.RoundTypeID == "c") {
				cpMap[fmt.Sprintf("%s-%s-%s", compId, result.EventID, result.PersonID)] = struct{}{}
			}

			addRWithCountry(result.RegionalSingleRecord)
			addRWithCountry(result.RegionalAverageRecord)

			if *result.PersonCountryID == country {
				personMap[result.PersonID] = struct{}{}
			}
		}
	}
	fs.A1Comps = len(competitions)
	fs.A2Rounds = len(roundMap)
	fs.A3Cp = len(cpMap)
	fs.B1Persons = len(personMap)

	var allPersonResultMap = make(map[string]struct{})
	var firstPersonResultMap = make(map[string]struct{})

	addRWithAll := func(record *string) {
		if record == nil {
			return
		}
		switch *record {
		case "WR":
			fs.C1WR += 1
		case "AsR":
			fs.C2AsR += 1
		case "NR":
			fs.C3NR += 1
		}
	}

	for _, result := range allWithCountryResult {
		addRWithAll(result.RegionalSingleRecord)
		addRWithAll(result.RegionalAverageRecord)
		allPersonResultMap[result.PersonID] = struct{}{}
		if strings.Contains(result.PersonID, fmt.Sprintf("%d", year)) {
			firstPersonResultMap[result.PersonID] = struct{}{}
		}
	}
	fs.B2Persons = len(allPersonResultMap)
	fs.B3Persons = len(firstPersonResultMap)

	tmpl, err := template.New("wcaStats").Parse(format)
	if err != nil {
		panic(err)
	}

	// 渲染模板
	bf := bytes.NewBuffer(nil)
	err = tmpl.Execute(bf, fs)
	if err != nil {
		panic(err)
	}

	return bf.String()
}
