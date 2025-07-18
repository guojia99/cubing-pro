package convenient

import (
	"context"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/configs"
	"github.com/guojia99/cubing-pro/src/internel/convenient/interface"
	"github.com/guojia99/cubing-pro/src/internel/convenient/job"
	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/crawler"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/sports"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	cache2 "github.com/patrickmn/go-cache"
	"gorm.io/gorm"
)

var _modelList []basemodel.DBModel

func Models() []interface{} {
	var out []interface{}
	for _, val := range _modelList {
		out = append(out, val)
	}
	return out
}

func NewConvenient(db *gorm.DB, runJob bool, config configs.Config) ConvenientI {
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
	_ = db.AutoMigrate(&competition.Registration{})                // 比赛注册表
	_ = db.AutoMigrate(&competition.AssCompetitionSponsorsUsers{}) // 比赛相关主办代表关联表
	_ = db.AutoMigrate(&competition.CompetitionGroup{})            // 比赛群组表

	// 爬虫表
	_ = db.AutoMigrate(&crawler.SendEmail{})

	// 运动表
	_ = db.AutoMigrate(&sports.SportEvent{})
	_ = db.AutoMigrate(&sports.SportResult{})

	// 系统
	_ = db.AutoMigrate(&system.KeyValue{}) // 系统数据表
	_ = db.AutoMigrate(&system.Image{})    // 系统图片表
	cache := cache2.New(time.Minute*5, time.Minute*5)

	out := &convenient{
		CompetitionIter: _interface.CompetitionIter{DB: db},
		UserIter:        _interface.UserIter{DB: db},
		ResultIter:      _interface.ResultIter{DB: db, Cache: cache},
		Jobs: []job.Job{
			{JobI: &job.RecordUpdateJob{DB: db}, Time: time.Minute * 30},
			//{JobI: &job.RecordUpdateJob{DB: db}, Time: time.Second * 3},
			{JobI: &job.UpdateDiyRankings{DB: db}, Time: time.Minute * 30},
		},
	}
	if runJob {
		out.Jobs.RunLoop(context.Background())
	}

	return out
}

type convenient struct {
	db *gorm.DB

	_interface.CompetitionIter
	_interface.UserIter
	_interface.ResultIter

	job.Jobs
}

func (c *convenient) DB() *gorm.DB { return c.db }

type ConvenientI interface {
	/*
		本接口目前有两种类型的
		Job: 定时任务
	*/
	DB() *gorm.DB
	_interface.CompetitionI
	_interface.UserI
	_interface.ResultI
}
