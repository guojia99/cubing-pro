package model

import "github.com/guojia99/cubing-pro/model/base"

var _modelList = []basemodel.DBModel{
	//&user.User{},               // 用户表
	//&UserMidea{},               // 用户媒体
	//&Organizers{},              // 主办团队表
	//&Event{},                   // 项目表
	//&Competition{},             // 比赛详情表
	//&CompetitionRegistration{}, // 比赛报名表
	//&CompetitionDiscussion{},   // 比赛直播讨论
	//&Results{},                 // 成绩表
	//&Record{},                  // 记录
	//&Notification{},            // 通知表
	//&PreResults{},              // 预录入成绩
	//&PostForm{},                // 交流文章表
	//&PostFormSub{},             // 交流消息
}

func Models() []basemodel.DBModel {
	return _modelList
}
