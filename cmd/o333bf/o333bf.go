package main

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/scramble"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db = "root@tcp(127.0.0.1:33306)/cubing_pro?charset=utf8&parseTime=True&loc=Local"
)

type Context struct {
	db *gorm.DB

	sb scramble.Scramble
}

func r1LoadDb(ctx *Context) (err error) {
	ctx.db, err = gorm.Open(mysql.New(mysql.Config{DSN: db}), &gorm.Config{})
	if err != nil {
		return err
	}

	ctx.sb = scramble.NewScramble(
		ctx.db,
		"rust_twisty",
		"",
		"",
		"",
	)

	return
}

func r2ResetAll333bfWithComp(ctx *Context) (err error) {
	var comps []competition.Competition

	_ = ctx.db.Find(&comps)

	for _, comp := range comps {
		if comp.IsDone {
			continue
		}
		for idx, ev := range comp.CompJSON.Events {
			if ev.EventID != "333bf" {
				continue
			}
			comp.CompJSON.Events[idx].EventRoute = event.RouteType5roundsBest
			for j, sd := range ev.Schedule {
				for k, s := range sd.Scrambles {
					if len(s) >= 7 {
						continue
					}
					sc := ctx.sb.Scramble("333bf", 2)

					comp.CompJSON.Events[idx].Schedule[j].Scrambles[k] = append(
						comp.CompJSON.Events[idx].Schedule[j].Scrambles[k],
						sc...,
					)
				}
			}
		}
		ctx.db.Save(&comp)
	}

	return
}

func r4ResetEvent(ctx *Context) (err error) {
	var ev event.Event
	if err = ctx.db.Where("id = ?", "333bf").First(&ev).Error; err != nil {
		return
	}
	ev.BaseRouteType = event.RouteType5roundsBest
	return ctx.db.Save(&ev).Error
}

func main() {
	builds := []func(ctx *Context) error{
		r1LoadDb,
		r2ResetAll333bfWithComp,
		r4ResetEvent, // 重置333bf
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
