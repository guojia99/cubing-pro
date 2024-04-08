package model

import (
	"github.com/guojia99/cubing-pro/backend/pkg/model/base"
	"github.com/guojia99/cubing-pro/backend/pkg/model/compertion"
	"github.com/guojia99/cubing-pro/backend/pkg/model/post"
	"github.com/guojia99/cubing-pro/backend/pkg/model/result"
	"github.com/guojia99/cubing-pro/backend/pkg/model/system"
	"github.com/guojia99/cubing-pro/backend/pkg/model/user"
)

var _modelList = []basemodel.DBModel{
	// 用户表
	&user.User{},
	&user.AuthRule{},
	&user.Role{},
	&user.Organizers{},
	&user.AssOrganizerUsers{},
	&user.AssUsersRoles{},
	&user.AssRoleAuthRule{},
	&user.UserKV{},

	// 讨论表和通知表
	&post.Forum{},
	&post.Topic{},
	&post.Posts{},
	&post.Notification{},
	&post.AssPostsLike{},
	&post.AssTopicLike{},

	//资源表
	&result.Event{},
	&result.Results{},
	&result.PreResults{},
	&result.Record{},

	//比赛表
	&compertion.Competition{},
	&compertion.CompetitionRegistration{},
	&compertion.AssCompetitionUsers{}, // 比赛相关主办代表关联表

	// 系统
	&system.KeyValue{},
	&system.Image{},
}

func Models() []interface{} {
	var out []interface{}
	for _, val := range _modelList {
		out = append(out, val)
	}
	return out
}
