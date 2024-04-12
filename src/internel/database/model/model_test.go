package model

import (
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
)

func _testDb(t *testing.T) *gorm.DB {
	db, err := gorm.Open(
		mysql.New(
			mysql.Config{
				DSN: "root:my123456@tcp(127.0.0.1:3306)/mycube3?charset=utf8&parseTime=True&loc=Local",
			},
		), &gorm.Config{
			Logger: logger.Discard,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestModels(t *testing.T) {
	db := _testDb(t)
	//
	//for _, val := range Models() {
	//	t.Log(reflect.TypeOf(val).Elem().Name())
	//	err := db.AutoMigrate(&val)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//}

	a := &user.AssUsersRoles{}
	if err := db.AutoMigrate(&a); err != nil {
		t.Fatal(err)
	}

}
