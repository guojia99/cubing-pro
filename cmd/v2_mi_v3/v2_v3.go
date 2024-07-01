package main

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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

	var ctx *Context
	for _, f := range builds {
		if err := f(ctx); err != nil {
			log.Fatalf("build failed error %s", err)
			return
		}
	}
	log.Println("ok")
}

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

	// v2 datas
	scores      []Score
	rounds      map[uint]Round
	contestList []Contest
	contests    map[uint]Contest
	PlayerList  []Player
	players     map[uint]Player
	playerUsers map[uint]PlayerUser
}

func r1LoadDb(ctx *Context) (err error) {
	ctx.v2Db, err = gorm.Open(mysql.New(mysql.Config{DSN: v2Db}), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		return err
	}
	ctx.v3Db, err = gorm.Open(mysql.New(mysql.Config{DSN: v2Db}), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		return err
	}
	return
}

func r2InitV2Datas(ctx *Context) (err error) {

	return
}

func r3ClearV3Datas(ctx *Context) (err error) {
	return
}

func r4InitV3BaseData(ctx *Context) (err error) {
	return
}

func r5SaveUser(ctx *Context) (err error) {
	return
}

func r6SaveV3CompetitionData(ctx *Context) (err error) {
	return
}

func r7SaveV3Results(ctx *Context) (err error) {
	return
}
