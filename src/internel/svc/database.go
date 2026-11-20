package svc

import (
	"github.com/guojia99/cubing-pro/src/configs"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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
				//Logger: dbLog,
			},
		)
	}
	return db, err
}
