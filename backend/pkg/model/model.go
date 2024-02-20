package model

import (
	"github.com/guojia99/cubing-pro/backend/pkg/model/base"
	"github.com/guojia99/cubing-pro/backend/pkg/model/compertion"
	"github.com/guojia99/cubing-pro/backend/pkg/model/result"
	"github.com/guojia99/cubing-pro/backend/pkg/model/user"
)

var _modelList = []basemodel.DBModel{
	// 用户表
	&user.AuthRule{},
	&user.User{},
	&user.Role{},
	&user.Organizers{},
	//资源表
	&result.Event{},
	&result.Results{},
	&result.PreResults{},
	&result.Record{},
	//比赛表
	&compertion.Competition{},
	&compertion.CompetitionRegistration{},
}

func Models() []interface{} {
	var out []interface{}
	for _, val := range _modelList {
		out = append(out, val)
	}
	return out
}
