package staticx

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type StaticX struct {
	db *gorm.DB
}

func (s *StaticX) Init() {
	dsn := "root@tcp(127.0.0.1:33306)/wca?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open(
		mysql.New(mysql.Config{DSN: dsn}), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		},
	)
	if err != nil {
		panic(err)
	}

	s.db = db
}
