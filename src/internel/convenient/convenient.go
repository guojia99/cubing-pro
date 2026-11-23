package convenient

import (
	"context"
	"time"

	"github.com/guojia99/cubing-pro/src/configs"
	"github.com/guojia99/cubing-pro/src/internel/convenient/interface"
	"github.com/guojia99/cubing-pro/src/internel/convenient/job"
	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
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
	migrate(db)

	cache := cache2.New(time.Minute*5, time.Minute*5)

	runJobs := newJobs(db, runJob, config)

	out := &convenient{
		CompetitionIter: _interface.CompetitionIter{DB: db},
		UserIter:        _interface.UserIter{DB: db},
		ResultIter:      _interface.ResultIter{DB: db, Cache: cache},
		Jobs:            runJobs,
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
	_interface.WCAResultI
}
