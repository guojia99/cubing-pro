package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/xuri/excelize/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ReadFileData() (bldData [][]string, otherData [][]string) {
	f, err := excelize.OpenFile("./大龄练习2025成绩2.xlsx")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	sheets := f.GetSheetList()

	// 第二个表为基准
	bldData, err = f.GetRows(sheets[1])
	if err != nil {
		panic(err)
	}

	otherData, err = f.GetRows(sheets[0])
	if err != nil {
		panic(err)
	}
	return
}

type CacheData struct {
	comp competition.Competition

	bldRoundNum map[string]int
	event       map[string]string
	register    map[string]competition.Registration
	results     []result.Results // name + compName + event + curRound
}

func getRound(isThree bool, idx int) string {
	round := "决赛"
	if isThree {
		switch idx {
		case 1:
			round = "初赛"
		case 2:
			round = "复赛"
		case 3:
			round = "决赛"
		}
	}
	return round
}

func newSchedule(isThree bool, idx int, evName string, roundNum int, startTime time.Time) competition.Schedule {
	return competition.Schedule{
		Round:           getRound(isThree, idx),
		Event:           evName,
		IsComp:          true,
		StartTime:       startTime,
		EndTime:         startTime.Add(24 * 7 * time.Hour),
		ActualStartTime: startTime,
		ActualEndTime:   startTime.Add(24 * 7 * time.Hour),
		RoundNum:        roundNum,
	}
}

// ParserToDbData
// events: key->cn
func ParserToDbData(curNewCompId uint, userMap map[string]user.User, events map[string]event.Event) ([]competition.Competition, []result.Results, []competition.Registration) {
	bldData, otherData := ReadFileData()

	var cache = make(map[string]*CacheData)

	// 盲
	for i := 1; i < len(bldData); i++ {
		// 新比赛
		compName := bldData[i][0]
		re := regexp.MustCompile(`\d{8}`)
		match := re.FindString(compName)
		startTime, _ := time.Parse("20060102", match)

		if _, ok := cache[compName]; !ok {
			curNewCompId += 1
			cache[compName] = &CacheData{
				register:    make(map[string]competition.Registration),
				event:       make(map[string]string),
				bldRoundNum: make(map[string]int),
				comp: competition.Competition{
					Model: basemodel.Model{
						ID:        curNewCompId,
						CreatedAt: startTime,
						UpdatedAt: time.Now(),
					},
					Status:  competition.Running,
					Name:    compName,
					Country: "中国",
					City:    "上海",
					CompJSON: competition.CompetitionJson{
						Events: make([]competition.CompetitionEvent, 0),
					},
					IsDone:        true,
					Genre:         competition.OnlineInformal,
					MinCount:      0,
					Count:         100,
					CompStartTime: startTime,
					CompEndTime:   startTime.Add(24 * 7 * time.Hour),
					OrganizersID:  4,
					GroupID:       4,
				},
				results: make([]result.Results, 0),
			}
			cache[compName].comp.CompJSON.Events = append(cache[compName].comp.CompJSON.Events, competition.CompetitionEvent{
				EventName:         "333bf",
				EventID:           "333bf",
				EventRoute:        event.RouteType3roundsBest,
				IsComp:            true,
				SingleQualify:     0,
				AvgQualify:        0,
				HasResultsQualify: false,
				Schedule: []competition.Schedule{
					newSchedule(true, 1, "333bf", 1, startTime),
					newSchedule(true, 2, "333bf", 2, startTime),
					newSchedule(true, 3, "333bf", 3, startTime),
				},
				Done: true,
			})
			// 添加赛程
		}

		userName := strings.ReplaceAll(bldData[i][2], " ", "")
		userId := userMap[userName].ID
		if userId == 0 {
			panic("not user" + userName)
		}
		var results []float64
		for _, res := range []string{bldData[i][3], bldData[i+1][3], bldData[i+2][3]} {
			results = append(results, result.TimeParserS2F(res))
		}

		if _, ok := cache[compName].register[userName]; !ok {
			cache[compName].register[userName] = competition.Registration{
				CompID:           cache[compName].comp.ID,
				CompName:         cache[compName].comp.Name,
				UserID:           userId,
				UserName:         userName,
				Status:           competition.RegisterStatusPass,
				RegistrationTime: time.Now(),
				AcceptationTime:  utils.PtrNow(),
				RetireTime:       nil,
				Events:           "",
				Payments:         nil,
				PaymentsJSON:     "",
			}
		}
		register := cache[compName].register[userName]
		register.SetEvent("333bf")
		cache[compName].register[userName] = register

		// 轮次
		roundNum, ok := cache[compName].bldRoundNum[userName]
		if !ok {
			cache[compName].bldRoundNum[userName] = 1
		}
		roundNum += 1
		cache[compName].bldRoundNum[userName] = roundNum

		newResult := result.Results{
			CompetitionID:   cache[compName].comp.ID, // 后面补充上去
			CompetitionName: compName,
			Round:           getRound(true, roundNum),
			RoundNumber:     roundNum,
			PersonName:      userName,
			UserID:          userId,
			CubeID:          userMap[userName].CubeID,
			Result:          results,
			EventID:         "333bf",
			EventName:       "333bf",
			EventRoute:      event.RouteType3roundsBest,
		}
		_ = newResult.Update()
		cache[compName].results = append(cache[compName].results, newResult)
		i += 2
	}

	// 其他
	for i := 1; i < len(otherData); i++ {
		compName := otherData[i][0]
		eventName := otherData[i][1]
		ev := events[eventName]

		// 判断这个项目需要加上去
		if _, ok := cache[compName].event[eventName]; !ok {
			cache[compName].comp.CompJSON.Events = append(cache[compName].comp.CompJSON.Events, competition.CompetitionEvent{
				EventName:  ev.Name,
				EventID:    ev.ID,
				EventRoute: ev.BaseRouteType,
				IsComp:     true,
				Schedule:   []competition.Schedule{newSchedule(false, 1, ev.Name, ev.BaseRouteType.RouteMap().Rounds, cache[compName].comp.CompStartTime)},
				Done:       true,
			})
			cache[compName].event[eventName] = ev.Name
		}

		userName := otherData[i][2]
		userId := userMap[userName].ID
		var results []float64
		for _, res := range otherData[i][3:] {
			results = append(results, result.TimeParserS2F(res))
		}
		//fmt.Println(otherData[i][3:], results)

		if _, ok := cache[compName].register[userName]; !ok {
			cache[compName].register[userName] = competition.Registration{
				CompID:           cache[compName].comp.ID,
				CompName:         cache[compName].comp.Name,
				UserID:           userId,
				UserName:         userName,
				Status:           competition.RegisterStatusPass,
				RegistrationTime: time.Now(),
				AcceptationTime:  utils.PtrNow(),
				RetireTime:       nil,
				Events:           "",
				Payments:         nil,
				PaymentsJSON:     "",
			}
		}
		register := cache[compName].register[userName]
		register.SetEvent(ev.ID)
		cache[compName].register[userName] = register

		newResult := result.Results{
			CompetitionID:   cache[compName].comp.ID, // 后面补充上去
			CompetitionName: compName,
			Round:           getRound(false, 1),
			RoundNumber:     1,
			PersonName:      userName,
			UserID:          userId,
			CubeID:          userMap[userName].CubeID,
			Result:          results,
			EventID:         ev.ID,
			EventName:       ev.Name,
			EventRoute:      ev.BaseRouteType,
		}
		_ = newResult.Update()
		cache[compName].results = append(cache[compName].results, newResult)
		//fmt.Println(newResult.PersonName, newResult.EventID, newResult.Best, newResult.Average)
	}

	var comps []competition.Competition
	var results []result.Results
	var registers []competition.Registration
	for _, c := range cache {
		comps = append(comps, c.comp)
		results = append(results, c.results...)
		for _, reg := range c.register {
			registers = append(registers, reg)
		}

	}

	sort.Slice(comps, func(i, j int) bool {
		return comps[i].ID < comps[j].ID
	})

	sort.Slice(results, func(i, j int) bool {
		if results[i].CompetitionID == results[j].CompetitionID {
			return results[i].RoundNumber < results[j].RoundNumber
		}
		return results[i].CompetitionID < results[j].CompetitionID
	})
	return comps, results, registers
}

func runParserToDb(db *gorm.DB) {
	// 基础数据
	var baseComp competition.Competition
	db.Order("id DESC").First(&baseComp)

	var events []event.Event
	var users []user.User
	db.Find(&events)
	db.Find(&users)
	var userMap = make(map[string]user.User)
	for _, usr := range users {
		userMap[usr.Name] = usr
	}
	var eventMap = make(map[string]event.Event)
	for _, ev := range events {
		eventMap[ev.ID] = ev
	}

	// 写入
	//ParserToDbData(baseComp.ID, userMap, eventMap)
	comps, results, reg := ParserToDbData(baseComp.ID, userMap, eventMap)
	db.Save(&comps)
	db.Save(&reg)
	db.Save(&results)
}

func main() {
	v3Db := "root@tcp(127.0.0.1:33306)/cubing_pro?charset=utf8&parseTime=True&loc=Local"
	//v3Db := "root:linwanting321_mysql_ttx1$%@tcp(127.0.0.1:3306)/cubing_pro?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: v3Db,
	}), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		panic(err)
	}

	// 删除大龄魔友盲第一场比赛
	//deleteID := 160
	//db.Delete(&competition.Competition{}, "id = ?", deleteID)
	//db.Delete(&competition.Registration{}, "comp_id = ?", deleteID)
	//db.Delete(&result.Results{}, "comp_id = ?", deleteID)

	// 删除之前的大龄比赛
	notDeleteName := "大龄盲拧周赛20250519第二十期"

	var findDelete []competition.Competition
	db.Unscoped().Where("name LIKE ?", "大龄盲拧周赛%").Find(&findDelete)

	for _, comp := range findDelete {
		if comp.Name == notDeleteName {
			continue
		}
		fmt.Println("delete -> ", comp.Name)
		// 生产
		//if comp.ID == 181{
		//	continue
		//}
		db.Unscoped().Delete(&comp, "id = ?", comp.ID)
	}

	// 录入前面的成绩
	runParserToDb(db)
}
