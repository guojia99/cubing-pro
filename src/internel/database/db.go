package database

import (
	"time"

	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"

	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
)

var _modelList = []basemodel.DBModel{
	// 用户表

}

func Models() []interface{} {
	var out []interface{}
	for _, val := range _modelList {
		out = append(out, val)
	}
	return out
}

func NewConvenient(db *gorm.DB) ConvenientI {
	_ = db.AutoMigrate()
	_ = db.AutoMigrate(&user.User{})       // 用户表
	_ = db.AutoMigrate(&user.CheckCode{})  // check code表
	_ = db.AutoMigrate(&user.Organizers{}) // 主办表
	_ = db.AutoMigrate(&user.UserKV{})     // 用户业务数据表

	// 讨论表和通知表
	_ = db.AutoMigrate(&post.Forum{})        // 板块表
	_ = db.AutoMigrate(&post.Topic{})        // 主题表
	_ = db.AutoMigrate(&post.Posts{})        // 恢复表
	_ = db.AutoMigrate(&post.Notification{}) // 通知表
	_ = db.AutoMigrate(&post.AssPostsLike{}) // post 点赞
	_ = db.AutoMigrate(&post.AssTopicLike{}) // 主题点赞

	//资源表
	_ = db.AutoMigrate(&event.Event{})       // 项目表
	_ = db.AutoMigrate(&result.Results{})    // 成绩表
	_ = db.AutoMigrate(&result.PreResults{}) // 预录入表
	_ = db.AutoMigrate(&result.Record{})     // 记录表

	//比赛表
	_ = db.AutoMigrate(&competition.Competition{})                 // 比赛表
	_ = db.AutoMigrate(&competition.CompetitionRegistration{})     // 比赛注册表
	_ = db.AutoMigrate(&competition.AssCompetitionSponsorsUsers{}) // 比赛相关主办代表关联表

	// 系统
	_ = db.AutoMigrate(&system.KeyValue{}) // 系统数据表
	_ = db.AutoMigrate(&system.Image{})    // 系统图片表

	return &convenient{
		db:    db,
		cache: cache.New(time.Minute*5, time.Minute*5),
	}
}

type convenient struct {
	db *gorm.DB

	cache *cache.Cache
}

func (c *convenient) DB() *gorm.DB { return c.db }

type ConvenientI interface {
	DB() *gorm.DB
	competitionI
	userI
	resultI
}
