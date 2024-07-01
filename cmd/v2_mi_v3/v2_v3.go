package main

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/guojia99/cubing-pro/cmd/v2_mi_v3/types"
	"github.com/guojia99/cubing-pro/src/internel/database"
	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	jsoniter "github.com/json-iterator/go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	v2Db = "root:my123456@tcp(127.0.0.1:3306)/mycube2?charset=utf8&parseTime=True&loc=Local"
	v3Db = "root:my123456@tcp(127.0.0.1:3306)/mycube3?charset=utf8&parseTime=True&loc=Local"
)

// 1. 将所有的数据拉到内存
// 2. 清空v3的数据库表
// 3. [载入] 生成默认主办团队, 添加默认项目表
// 4. [载入] 添加用户数据, 生成映射表
// 5. [载入] 添加比赛数据，添加比赛注册数据, 生成映射表
// 6. [载入] 按指定项目载入成绩表。

type Context struct {
	v2Db *gorm.DB
	v3Db *gorm.DB

	evSort map[string]int

	// v2 datas
	scoresList []types.Score

	roundsList []types.Round
	rounds     map[uint]types.Round

	contestList []types.Contest
	contests    map[uint]types.Contest

	PlayerList []types.Player
	players    map[uint]types.Player

	playerUserList []types.PlayerUser
	playerUsers    map[uint]types.PlayerUser

	// v3 datas
	it       database.ConvenientI
	V3events map[string]event.Event
	V3Users  map[uint]user.User               // 这里用的是v2的Id作为key
	V3Comps  map[uint]competition.Competition // 原本的比赛ID
	v3Org    map[string]user.Organizers       // 魔缘 or 盲拧
}

func r1LoadDb(ctx *Context) (err error) {
	ctx.v2Db, err = gorm.Open(mysql.New(mysql.Config{DSN: v2Db}), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		return err
	}
	ctx.v3Db, err = gorm.Open(mysql.New(mysql.Config{DSN: v3Db}), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		return err
	}
	return
}

func r2InitV2Datas(ctx *Context) (err error) {
	ctx.v2Db.Find(&ctx.scoresList)
	ctx.v2Db.Find(&ctx.roundsList)
	ctx.v2Db.Find(&ctx.contestList)
	ctx.v2Db.Find(&ctx.PlayerList)
	ctx.v2Db.Find(&ctx.playerUserList)

	ctx.contests = make(map[uint]types.Contest)
	ctx.players = make(map[uint]types.Player)
	ctx.playerUsers = make(map[uint]types.PlayerUser)
	ctx.rounds = make(map[uint]types.Round)

	for _, val := range ctx.roundsList {
		ctx.rounds[val.ID] = val
	}
	for _, val := range ctx.contestList {
		ctx.contests[val.ID] = val
	}
	for _, val := range ctx.playerUserList {
		ctx.playerUsers[val.PlayerID] = val
	}
	for _, val := range ctx.PlayerList {
		ctx.players[val.ID] = val
	}
	return
}

func r3ClearV3Datas(ctx *Context) (err error) {
	ctx.it = database.NewConvenient(ctx.v3Db)

	tables := []interface{}{
		&user.User{},
		&user.Organizers{},
		&event.Event{},
		&result.Results{},
		&competition.Competition{},
		&competition.CompetitionRegistration{},
	}

	for _, t := range tables {
		if err = ctx.v3Db.Unscoped().Delete(&t, "1 = 1").Error; err != nil {
			return err
		}
	}
	return
}

func r4InitV3BaseData(ctx *Context) (err error) {
	ctx.v3Db.Save(&types.AllEvents)
	ctx.V3events = make(map[string]event.Event)

	ctx.evSort = make(map[string]int)
	for i, e := range types.AllEvents {
		ctx.evSort[e.ID] = i
	}

	var events []event.Event
	ctx.v3Db.Find(&events)
	for _, ev := range events {
		ctx.V3events[ev.ID] = ev
	}
	return
}

func r5SaveUser(ctx *Context) (err error) {
	ctx.V3Users = make(map[uint]user.User)
	ctx.v3Org = make(map[string]user.Organizers)
	for _, val := range ctx.PlayerList {
		u := ctx.players[val.ID]
		usr, _ := ctx.playerUsers[val.ID]

		newUser := user.User{
			Model: basemodel.Model{
				ID:        u.ID,
				CreatedAt: u.CreatedAt,
				UpdatedAt: u.CreatedAt,
			},
			Auth:         user.AuthPlayer,
			Name:         u.Name,
			Hash:         uuid.NewString(),
			InitPassword: uuid.NewString(),
			Level:        3,
			Experience:   10000,
			QQ:           usr.QQ,
			QQUniID:      usr.QQBotUniID,
			Wechat:       usr.WeChat,
			WcaID:        u.WcaID,
			Phone:        usr.Phone,
			ActualName:   u.ActualName,
			CubeID:       ctx.it.GetCubeID(u.Name),
		}
		if newUser.Name == "嘉吖" {
			newUser.SetAuth(user.AuthOrganizers, user.AuthAdmin, user.AuthSuperAdmin)
		}
		if newUser.Name == "模仿者Wing" || newUser.Name == "小丫鬟" {
			newUser.SetAuth(user.AuthOrganizers)

			// 创建主办团队
			var org user.Organizers
			switch newUser.Name {
			case "模仿者Wing":
				org = user.Organizers{
					Name:         "中国盲拧战队",
					Introduction: "中国盲拧战队群赛",
					Email:        "chinabf@gmail.com",
					QQGroup:      "941777598,942909225",
					QQGroupUid:   "BF9E9681703B83E5A5626831756E5977,A46A01E1E5F7D3B8980BCDB6FF868717",
					LeaderID:     newUser.CubeID,
					Status:       user.Using,
				}
			case "小丫鬟":
				org = user.Organizers{
					Model:        basemodel.Model{},
					Name:         "魔缘群",
					Introduction: "磨圆群",
					Email:        "moyuan@gmail.com",
					QQGroup:      "563250032",
					QQGroupUid:   "B613DB043FBAF68F73BB915F98E61BF3",
					LeaderID:     newUser.CubeID,
					Status:       user.Using,
				}
			}
			if err = ctx.v3Db.Create(&org).Error; err != nil {
				return err
			}
			ctx.v3Org[org.Name] = org
		}

		ctx.V3Users[val.ID] = newUser
		if err = ctx.v3Db.Model(&user.User{}).Create(&newUser).Error; err != nil {
			return err
		}
	}
	return
}

func _cutBfGroupRound(ctx *Context, in string) (ev event.Event, groupNum int) {
	cut0, cut1 := in[:len(in)-1], string(in[len(in)-1])
	groupNum, _ = strconv.Atoi(cut1)

	switch cut0 {
	case "3bf":
		ev = ctx.V3events["333bf"]
	case "4bf":
		ev = ctx.V3events["444bf"]
	case "5bf":
		ev = ctx.V3events["555bf"]
	case "3mbf":
		ev = ctx.V3events["333mbf"]
	}
	return
}

func r6SaveV3CompetitionData(ctx *Context) (err error) {
	// 盲拧战队群的比赛有多轮的，需要结合成一轮
	var comps []competition.Competition
	for _, c := range ctx.contestList {
		if !strings.Contains(c.Name, "魔缘") && !strings.Contains(c.Name, "盲拧战队") {
			continue
		}

		key := "魔缘群"
		if strings.Contains(c.Name, "盲拧战队") {
			key = "中国盲拧战队"
		}

		newComp := competition.Competition{
			Model: basemodel.Model{
				ID:        c.ID,
				CreatedAt: c.CreatedAt,
				UpdatedAt: c.CreatedAt,
			},
			StrId:      "",
			Status:     competition.Running,
			RejectMsg:  "",
			Name:       c.Name,
			Illustrate: fmt.Sprintf("## %s(v2版本)", c.Name),
			City:       "guangzhou",
			RuleMD:     "群赛",
			Series:     "",
			Genre: func() competition.Genre {
				if c.Type == "offline" {
					return competition.Informal
				}
				return competition.OnlineInformal
			}(),
			MinCount:           0,
			Count:              1000,
			FreeParticipate:    false,
			AutomaticReview:    true,
			CanPreResult:       true,
			CanStartedAddEvent: false,
			CompStartTime:      c.StartTime,
			CompEndTime:        c.EndTime,
			OrganizersID:       ctx.v3Org[key].ID,
		}
		var roundNums []uint
		_ = jsoniter.UnmarshalFromString(c.RoundIds, &roundNums)
		newComp.CompJSON = competition.CompetitionJson{
			Events: make([]competition.CompetitionEvent, 0),
		}
		var events = make(map[string]competition.CompetitionEvent)

		for _, roundNum := range roundNums {
			round := ctx.rounds[roundNum]

			var ev event.Event
			switch key {
			case "中国盲拧战队":
				ev, round.Number = _cutBfGroupRound(ctx, round.Project)
			default:
				var ok bool
				ev, ok = ctx.V3events[round.Project]
				if !ok {
					continue
				}
			}
			var cEvent = competition.CompetitionEvent{
				EventName:         ev.Name,
				EventID:           ev.ID,
				EventRoute:        ev.BaseRouteType,
				IsComp:            ev.IsComp,
				SingleQualify:     0,
				AvgQualify:        0,
				HasResultsQualify: false,
				Schedule:          make([]competition.Schedule, 0),
				Done:              true,
			}
			if evs, ok := events[ev.ID]; ok {
				cEvent = evs
			}
			cEvent.Schedule = append(
				cEvent.Schedule,
				competition.Schedule{
					Round:           round.Name,
					Event:           ev.Name,
					IsComp:          ev.IsComp,
					StartTime:       newComp.CompStartTime,
					EndTime:         newComp.CompEndTime,
					ActualStartTime: newComp.CompStartTime,
					ActualEndTime:   newComp.CompEndTime,
					RoundNum:        round.Number,
					IsRunning:       false,
				},
			)
			events[ev.ID] = cEvent
		}
		for _, ev := range events {
			newComp.CompJSON.Events = append(newComp.CompJSON.Events, ev)
		}
		sort.Slice(
			newComp.CompJSON.Events, func(i, j int) bool {
				return ctx.evSort[newComp.CompJSON.Events[i].EventID] < ctx.evSort[newComp.CompJSON.Events[j].EventID]
			},
		)
		var eventList = func() []string {
			var out []string
			for _, val := range newComp.CompJSON.Events {
				out = append(out, val.EventID)
			}
			return out
		}()
		newComp.EventMin = strings.Join(eventList, ",")
		if len(newComp.CompJSON.Events) == 0 {
			//fmt.Printf("移除所有 %s 比赛成绩\n", newComp.Name)
			continue
		}
		//fmt.Println(newComp.Name, newComp.EventMin)
		comps = append(comps, newComp)
	}
	if err = ctx.v3Db.Create(&comps).Error; err != nil {
		return err
	}

	var all []competition.Competition
	ctx.v3Db.Find(&all)
	ctx.V3Comps = make(map[uint]competition.Competition)
	for _, comp := range all {
		ctx.V3Comps[comp.ID] = comp
	}

	return
}

func _getResultEv(ctx *Context, project string) (event.Event, int, error) {
	if ev, ok := ctx.V3events[project]; ok {
		return ev, -1, nil
	}

	ev, number := _cutBfGroupRound(ctx, project)
	if ev.Name == "" {
		return ev, -1, errors.New("error")
	}
	return ev, number, nil
}

func r7SaveV3Results(ctx *Context) (err error) {
	// 循环所有成绩， 如果查看该用户是否注册， 是否已经添加到map
	var regs = make(map[string]competition.CompetitionRegistration)
	var results []result.Results

	for _, score := range ctx.scoresList {
		ev, number, err := _getResultEv(ctx, score.Project)
		if err != nil {
			continue
		}

		round := ctx.rounds[score.RouteID]

		// 确认是否已经加过比赛
		key := fmt.Sprintf("%d-%d", score.PlayerID, score.ContestID)
		reg, ok := regs[key]
		if !ok {
			reg = competition.CompetitionRegistration{
				CompID:           ctx.V3Comps[score.ContestID].ID,
				CompName:         ctx.V3Comps[score.ContestID].Name,
				UserID:           ctx.V3Users[score.PlayerID].ID,
				UserName:         ctx.V3Users[score.PlayerID].Name,
				Status:           competition.RegisterStatusPass,
				RegistrationTime: score.CreatedAt,
				AcceptationTime:  utils.PtrTime(score.CreatedAt),
				Events:           "",
			}
		}
		reg.Events += fmt.Sprintf(",%s", ev.ID)
		regs[key] = reg

		// 添加比赛成绩
		newResult := result.Results{
			Model: basemodel.Model{
				CreatedAt: score.CreatedAt,
				UpdatedAt: score.CreatedAt,
			},
			CompetitionID: ctx.V3Comps[score.ContestID].ID,
			Round:         round.Name,
			RoundNumber:   number,
			PersonName:    ctx.V3Users[score.PlayerID].Name,
			UserID:        ctx.V3Users[score.PlayerID].ID,
			Result: []float64{
				score.Result1, score.Result2, score.Result3, score.Result4, score.Result5,
			},
			EventID:    ev.ID,
			EventName:  ev.Name,
			EventRoute: ev.BaseRouteType,
		}
		if err = newResult.Update(); err != nil {
			return err
		}

		if newResult.EventRoute.RouteMap().Repeatedly && newResult.EventID == "333mbf" && newResult.BestRepeatedlyTime > 3700 {
			newResult.EventID = "333mbf_unlimited"
		}

		results = append(results, newResult)
	}

	for i := 0; i < len(results); i += 100 {
		end := i + 100
		if end > len(results) {
			end = len(results)
		}
		list := results[i:end]
		if err = ctx.v3Db.Create(&list).Error; err != nil {
			return err
		}
	}

	var regList []competition.CompetitionRegistration
	for _, reg := range regs {
		s := strings.Split(reg.Events, ",")
		s = slices.DeleteFunc(s, func(s string) bool { return !(s != " " && s != "") })
		s = utils.RemoveRepeatedElement(s)

		sort.Slice(s, func(i, j int) bool { return ctx.evSort[s[i]] < ctx.evSort[s[j]] })

		reg.Events, _ = jsoniter.MarshalToString(s)
		regList = append(regList, reg)
	}

	if err = ctx.v3Db.Create(&regList).Error; err != nil {
		return err
	}
	return
}

func main() {
	builds := []func(ctx *Context) error{
		r1LoadDb,
		r2InitV2Datas,
		r3ClearV3Datas,
		r4InitV3BaseData,
		r5SaveUser,
		r6SaveV3CompetitionData,
		r7SaveV3Results,
	}

	var ctx = &Context{}
	for _, f := range builds {
		ts := time.Now()
		if err := f(ctx); err != nil {
			log.Fatalf("build failed error %s", err)
			return
		}
		ftyp := reflect.TypeOf(f)
		fmt.Printf("%s use time %s\n", ftyp.String(), time.Since(ts))
	}
	log.Println("ok")
}
