package svc

import (
	"fmt"

	"github.com/guojia99/cubing-pro/src/configs"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newDB(cfg configs.GlobalConfig) (*gorm.DB, error) {
	if cfg.DB.Driver != "mysql" {
		return nil, fmt.Errorf("unsupported database driver %q: only mysql is supported", cfg.DB.Driver)
	}
	dbLog := logger.Discard
	if cfg.Debug {
		dbLog = logger.Default.LogMode(logger.Info)
	}
	return gorm.Open(mysql.Open(cfg.DB.DSN), &gorm.Config{Logger: dbLog})
}
