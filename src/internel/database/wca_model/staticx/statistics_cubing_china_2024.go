package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Result struct {
	CompetitionID         string  `gorm:"column:competitionId; type:varchar(32); not null; default:''" json:"competitionId"`
	EventID               string  `gorm:"column:eventId; type:varchar(6); not null; default:''" json:"eventId"`
	RoundTypeID           string  `gorm:"column:roundTypeId; type:varchar(1); not null; default:''" json:"roundTypeId"`
	Pos                   int16   `gorm:"column:pos; type:smallint; not null; default:0" json:"pos"`
	Best                  int     `gorm:"column:best; type:int; not null; default:0" json:"best"`
	Average               int     `gorm:"column:average; type:int; not null; default:0" json:"average"`
	PersonName            *string `gorm:"column:personName; type:varchar(80)" json:"personName,omitempty"`
	PersonID              string  `gorm:"column:personId; type:varchar(10); not null; default:''" json:"personId"`
	PersonCountryID       *string `gorm:"column:personCountryId; type:varchar(50)" json:"personCountryId,omitempty"`
	FormatID              string  `gorm:"column:formatId; type:varchar(1); not null; default:''" json:"formatId"`
	Value1                int     `gorm:"column:value1; type:int; not null; default:0" json:"value1"`
	Value2                int     `gorm:"column:value2; type:int; not null; default:0" json:"value2"`
	Value3                int     `gorm:"column:value3; type:int; not null; default:0" json:"value3"`
	Value4                int     `gorm:"column:value4; type:int; not null; default:0" json:"value4"`
	Value5                int     `gorm:"column:value5; type:int; not null; default:0" json:"value5"`
	RegionalSingleRecord  *string `gorm:"column:regionalSingleRecord; type:varchar(3)" json:"regionalSingleRecord,omitempty"`
	RegionalAverageRecord *string `gorm:"column:regionalAverageRecord; type:varchar(3)" json:"regionalAverageRecord,omitempty"`
}

func (Result) TableName() string {
	return "Results"
}

type Competition struct {
	ID              string  `gorm:"column:id; type:varchar(32); not null; default:''" json:"id"`
	Name            string  `gorm:"column:name; type:varchar(50); not null; default:''" json:"name"`
	CityName        string  `gorm:"column:cityName; type:varchar(50); not null; default:''" json:"cityName"`
	CountryID       string  `gorm:"column:countryId; type:varchar(50); not null; default:''" json:"countryId"`
	Information     *string `gorm:"column:information; type:mediumtext" json:"information,omitempty"`
	Year            uint16  `gorm:"column:year; type:smallint unsigned; not null; default:0" json:"year"`
	Month           uint16  `gorm:"column:month; type:smallint unsigned; not null; default:0" json:"month"`
	Day             uint16  `gorm:"column:day; type:smallint unsigned; not null; default:0" json:"day"`
	EndMonth        uint16  `gorm:"column:endMonth; type:smallint unsigned; not null; default:0" json:"endMonth"`
	EndDay          uint16  `gorm:"column:endDay; type:smallint unsigned; not null; default:0" json:"endDay"`
	Cancelled       int     `gorm:"column:cancelled; type:int; not null; default:0" json:"cancelled"`
	EventSpecs      *string `gorm:"column:eventSpecs; type:longtext" json:"eventSpecs,omitempty"`
	WcaDelegate     *string `gorm:"column:wcaDelegate; type:mediumtext" json:"wcaDelegate,omitempty"`
	Organiser       *string `gorm:"column:organiser; type:mediumtext" json:"organiser,omitempty"`
	Venue           string  `gorm:"column:venue; type:varchar(240); not null; default:''" json:"venue"`
	VenueAddress    *string `gorm:"column:venueAddress; type:varchar(191)" json:"venueAddress,omitempty"`
	VenueDetails    *string `gorm:"column:venueDetails; type:varchar(191)" json:"venueDetails,omitempty"`
	ExternalWebsite *string `gorm:"column:external_website; type:varchar(200)" json:"externalWebsite,omitempty"`
	CellName        string  `gorm:"column:cellName; type:varchar(45); not null; default:''" json:"cellName"`
	Latitude        *int    `gorm:"column:latitude; type:int" json:"latitude,omitempty"`
	Longitude       *int    `gorm:"column:longitude; type:int" json:"longitude,omitempty"`
}

func (Competition) TableName() string {
	return "Competitions"
}

type StaticX struct {
	db *gorm.DB
}

func (s *StaticX) Init() {
	dsn := "root@tcp(127.0.0.1:33306)/wca_dev?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open(
		mysql.New(mysql.Config{DSN: dsn}), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		},
	)
	if err != nil {
		panic(err)
	}

	s.db = db
}

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

func main() {
	s := StaticX{}
	s.Init()
	//s.BaseData(false)
	data := s.GetStatic("China", 2024)
	fmt.Println(data)
	//data2 := s.GetStatic("Hong Kong", 2024)
	//fmt.Println(data2)
	//data3 := s.GetStatic("TaiWan", 2024)
	//fmt.Println(data3)
	//data4 := s.GetStatic("Macau", 2024)
	//fmt.Println(data4)
}
