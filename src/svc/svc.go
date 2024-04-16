/*
 *  * Copyright (c) 2023 guojia99 All rights reserved.
 *  * Created: 2023/7/10 下午2:43.
 *  * Author: guojia(https://github.com/guojia99)
 */

package svc

import (
	"time"

	"github.com/patrickmn/go-cache"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/guojia99/cubing-pro/src/internel/database"
)

type Svc struct {
	DB    *gorm.DB
	Cache *cache.Cache
	Cfg   Config
	Cov   database.ConvenientI
}

func NewAPISvc(file string) (*Svc, error) {
	var err error
	var cfg Config
	if err = cfg.Load(file); err != nil {
		return nil, err
	}

	c := &Svc{
		Cfg:   cfg,
		Cache: cache.New(time.Minute*5, time.Minute*5),
	}
	if c.DB, err = newDB(cfg.GlobalConfig.DB); err != nil {
		return nil, err
	}
	c.Cov = database.NewConvenient(c.DB)
	return c, nil
}

func newDB(cfg DBConfig) (*gorm.DB, error) {
	var err error
	var db *gorm.DB
	switch cfg.Driver {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{})
	case "mysql":
		db, err = gorm.Open(
			mysql.New(mysql.Config{DSN: cfg.DSN}), &gorm.Config{
				Logger: logger.Discard,
			},
		)
	}
	return db, err
}
