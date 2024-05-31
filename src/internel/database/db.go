package database

import (
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
	_ = db.AutoMigrate(&user.User{})
	_ = db.AutoMigrate(&user.CheckCode{})
	_ = db.AutoMigrate(&user.Organizers{})
	_ = db.AutoMigrate(&user.UserKV{})

	// 讨论表和通知表
	_ = db.AutoMigrate(&post.Forum{})
	_ = db.AutoMigrate(&post.Topic{})
	_ = db.AutoMigrate(&post.Posts{})
	_ = db.AutoMigrate(&post.Notification{})
	_ = db.AutoMigrate(&post.AssPostsLike{})
	_ = db.AutoMigrate(&post.AssTopicLike{})

	//资源表
	_ = db.AutoMigrate(&event.Event{})
	_ = db.AutoMigrate(&result.Results{})
	_ = db.AutoMigrate(&result.PreResults{})
	_ = db.AutoMigrate(&result.Record{})

	//比赛表
	_ = db.AutoMigrate(&competition.Competition{})
	_ = db.AutoMigrate(&competition.CompetitionRegistration{})
	_ = db.AutoMigrate(&competition.AssCompetitionSponsorsUsers{}) // 比赛相关主办代表关联表

	// 系统
	_ = db.AutoMigrate(&system.KeyValue{})
	_ = db.AutoMigrate(&system.Image{})

	return &convenient{db: db}
}

type convenient struct {
	db *gorm.DB
}

func (c *convenient) DB() *gorm.DB { return c.db }

type ConvenientI interface {
	DB() *gorm.DB
	competitionI
	userI
}
