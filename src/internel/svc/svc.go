/*
 *  * Copyright (c) 2023 guojia99 All rights reserved.
 *  * Created: 2023/7/10 下午2:43.
 *  * Author: guojia(https://github.com/guojia99)
 */

package svc

import (
	"time"

	"github.com/guojia99/cubing-pro/src/configs"
	"github.com/guojia99/cubing-pro/src/internel/convenient"
	"github.com/guojia99/cubing-pro/src/internel/scramble"
	"gorm.io/gorm/logger"

	"github.com/patrickmn/go-cache"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Svc struct {
	DB       *gorm.DB
	Cache    *cache.Cache
	Cfg      configs.Config
	Cov      convenient.ConvenientI
	Scramble scramble.Scramble
}

func NewAPISvc(file string, job bool, scr bool) (*Svc, error) {
	var err error
	var cfg configs.Config
	if err = cfg.Load(file); err != nil {
		return nil, err
	}

	c := &Svc{
		Cfg:   cfg,
		Cache: cache.New(time.Minute*5, time.Minute*5),
	}
	if scr {
		c.Scramble = scramble.NewScramble(cfg.GlobalConfig.Scramble.Type, cfg.GlobalConfig.Scramble.EndPoint, cfg.GlobalConfig.Scramble.ScrambleDrawType, cfg.GlobalConfig.Scramble.ScrambleUrl)
	}

	if c.DB, err = newDB(cfg.GlobalConfig); err != nil {
		return nil, err
	}

	// todo 多个程序时
	c.Cov = convenient.NewConvenient(c.DB, job, cfg)
	return c, nil
}

func newDB(cfg configs.GlobalConfig) (*gorm.DB, error) {
	var err error
	var db *gorm.DB

	var dbLog = logger.Discard

	if cfg.Debug {
		dbLog = logger.Default.LogMode(logger.Info) // 将日志模式设置为 Info
	}

	switch cfg.DB.Driver {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.DB.DSN), &gorm.Config{Logger: dbLog})
	case "mysql":
		db, err = gorm.Open(
			mysql.New(mysql.Config{DSN: cfg.DB.DSN}), &gorm.Config{
				Logger: dbLog,
			},
		)
	}
	return db, err
}
