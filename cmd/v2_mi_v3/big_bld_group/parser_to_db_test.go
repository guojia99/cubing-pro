package main

import (
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Test_runParserToDb(t *testing.T) {
	v3Db := "root@tcp(127.0.0.1:33306)/cubing_pro?charset=utf8&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: v3Db,
	}), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		panic(err)
	}
	runParserToDb(db)
}
