package database

import (
	"fmt"
	"os"
	"reflect"
	"slices"
	"sort"
	"testing"

	"github.com/gookit/color"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/go-tables/table"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var wca = []string{
	"222",
	"333",
	"333bf",
	"333fm",
	"333mbf",
	"333oh",
	"444",
	"444bf",
	"555",
	"555bf",
	"666",
	"777",
	"clock",
	"minx",
	"pyram",
	"skewb",
	"sq1",
	//
	//"555",
	//"666",
	//"777",
	//"333fm",

	//"333bf",
	//"333mbf",
	//"444bf",
	//"555bf",

	//"222",
	//"333",
	//"444",
	//"555",
	//"666",
	//"777",
	//"333oh",

	//"clock",
	//"minx",
	//"pyram",
	//"skewb",
	//"sq1",
}

func Test_convenient_KinChSor(t *testing.T) {
	var v3Db = "root:my123456@tcp(127.0.0.1:3306)/mycube3?charset=utf8&parseTime=True&loc=Local"
	Db, err := gorm.Open(mysql.New(mysql.Config{DSN: v3Db}), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	c := NewConvenient(Db)

	var results []result.Results
	var userIds []uint
	Db.Where("event_id in ?", wca).Find(&results)
	Db.Model(&result.Results{}).Distinct("user_id").Where("event_id in ?", wca).Find(&userIds)
	var events []event.Event
	Db.Where("id in ?", wca).Find(&events)

	var players []user.User
	Db.Where("id in ?", userIds).Find(&players)

	best, all := c.AllPlayerBestResult(results, players)
	kr := c.KinChSor(best, events, all)

	tb := table.NewTable(table.DefaultOption)
	{
		h := []interface{}{"玩家", "分数"}
		for _, e := range events {
			h = append(h, e.ID)
		}
		tb.SetHeaders(h...)
	}

	for idx, r := range kr {
		var body []interface{}
		body = append(body, r.PlayerName)
		body = append(body, fmt.Sprintf("%.2f", r.Result))
		for _, rs := range r.Results {
			body = append(body, fmt.Sprintf("%.2f", rs.Result))
		}
		tb.AddBody(body...)
		if idx > 50 {
			break
		}
	}

	tb.Opt.TransformContents = append(
		tb.Opt.TransformContents, func(in interface{}) interface{} {
			if in == "100.00" {
				return color.Danger.Sprintf("%v", in)
			}
			if in == "0.00" {
				return color.Warn.Sprintf("%v", in)
			}
			return in
		},
	)
	os.WriteFile("test.txt", []byte(tb.String()), 0644)
}

func Test_convenient_PlayerNemesis(t *testing.T) {
	var v3Db = "root:my123456@tcp(127.0.0.1:3306)/mycube3?charset=utf8&parseTime=True&loc=Local"
	Db, err := gorm.Open(mysql.New(mysql.Config{DSN: v3Db}), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	c := NewConvenient(Db)

	var results []result.Results
	var userIds []uint
	Db.Where("event_id in ?", wca).Find(&results)
	Db.Model(&result.Results{}).Distinct("user_id").Where("event_id in ?", wca).Find(&userIds)

	var events []event.Event
	Db.Where("id in ?", wca).Find(&events)
	var eventMap = make(map[string]event.Event)
	for _, e := range events {
		eventMap[e.ID] = e
	}

	var players []user.User
	Db.Where("id in ?", userIds).Find(&players)

	_, all := c.AllPlayerBestResult(results, players)

	tb := table.NewTable(table.DefaultOption)
	tb.SetHeaders("名字", "宿敌数(全项目)", "宿敌数(仅双方都有的项目)")

	var bodys [][]interface{}
	for _, p := range all {
		nemesis1 := c.PlayerNemesis(p, all, eventMap, false)
		nemesis2 := c.PlayerNemesis(p, all, eventMap, true)

		bodys = append(
			bodys, []interface{}{
				p.PlayerName,
				len(nemesis1),
				len(nemesis2),
			},
		)
	}
	sort.Slice(
		bodys, func(i, j int) bool {
			if bodys[i][1].(int) == bodys[j][1].(int) {
				return bodys[i][2].(int) < bodys[j][2].(int)
			}

			return bodys[i][1].(int) < bodys[j][1].(int)
		},
	)

	tb.Opt.TransformContents = append(
		tb.Opt.TransformContents, func(in interface{}) interface{} {
			if in == 0 {
				return color.Green.Sprintf("%v", in)
			}
			typ := reflect.TypeOf(in)
			if typ.Kind() != reflect.Int {
				return in
			}
			val := reflect.ValueOf(in)
			if val.Int() < 5 {
				return color.Warn.Sprintf("%v", in)
			}
			return in
		},
	)

	for _, b := range bodys {
		tb.AddBody(b...)
	}
	os.WriteFile("test.txt", []byte(tb.String()), 0644)
}

func Test_convenient_AllPlayerBestResult(t *testing.T) {
	var v3Db = "root:my123456@tcp(127.0.0.1:3306)/mycube3?charset=utf8&parseTime=True&loc=Local"
	Db, err := gorm.Open(mysql.New(mysql.Config{DSN: v3Db}), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	c := NewConvenient(Db)

	var results []result.Results
	var userIds []uint
	Db.Where("event_id in ?", wca).Find(&results)
	Db.Model(&result.Results{}).Distinct("user_id").Where("event_id in ?", wca).Find(&userIds)
	var players []user.User
	Db.Find(&players)

	_, all := c.AllPlayerBestResult(results, players)
	names := []string{"冰渊", "小丫鬟"}
	for _, val := range all {
		if slices.Contains(names, val.PlayerName) {
			tb := table.NewTable(table.DefaultOption)
			tb.SetHeaders("项目", "单次", "平均")
			for _, ev := range wca {
				s, ok1 := val.Single[ev]
				a, ok2 := val.Avgs[ev]

				if !ok1 {
					continue
				}

				var body []interface{}
				body = append(body, ev)
				body = append(body, s.BestString())

				if ok2 {
					body = append(body, a.BestAvgString())
				} else {
					body = append(body, " ")
				}
				tb.AddBody(body...)
			}
			fmt.Printf("---------- %s ----------\n", val.PlayerName)
			fmt.Println(tb)
			continue
		}
	}

}
