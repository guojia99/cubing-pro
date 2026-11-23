package convenient

import (
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/crawler"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/pktimer"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/sports"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	wca_model "github.com/guojia99/cubing-pro/src/internel/database/model/wca"
	"gorm.io/gorm"
)

func migrate(db *gorm.DB) {
	_ = db.AutoMigrate()
	_ = db.AutoMigrate(&user.User{})       // 用户表
	_ = db.AutoMigrate(&user.OAuthState{}) // OAuth state 表（WCA 登录防 CSRF）
	_ = db.AutoMigrate(&user.CheckCode{})  // check code表
	_ = db.AutoMigrate(&user.Organizers{}) // 主办表
	_ = db.AutoMigrate(&user.UserKV{})     // 用户业务数据表

	// 讨论表和通知表
	_ = db.AutoMigrate(&post.Forum{})        // 板块表
	_ = db.AutoMigrate(&post.Topic{})        // 主题表
	_ = db.AutoMigrate(&post.Posts{})        // 回复表
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

	// wca
	_ = db.AutoMigrate(&wca_model.WCAResult{}) // wca 信息缓存表

	_ = db.AutoMigrate(&pktimer.PkTimerResult{}) // pk表
}
