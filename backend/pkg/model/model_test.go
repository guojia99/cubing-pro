package model

import (
	"reflect"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func _testDb() *gorm.DB {
	db, _ := gorm.Open(
		mysql.New(
			mysql.Config{
				DSN: "root:my123456@tcp(127.0.0.1:3306)/mycube3?charset=utf8&parseTime=True&loc=Local",
			},
		), &gorm.Config{
			Logger: logger.Discard,
		},
	)
	return db
}

func TestModels(t *testing.T) {
	db := _testDb()

	for _, val := range Models() {
		err := db.AutoMigrate(val)
		if err != nil {
			t.Log(reflect.TypeOf(val).Elem().Name())
			t.Fatal(err)
		}
	}
}
