package _interface

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//func TestResultIter_SelectSeniorKinChSor(t *testing.T) {
//
//	c := &ResultIter{
//		Cache: cache.New(time.Minute, time.Minute),
//	}
//
//
//	c.SelectSeniorKinChSor(1, 20, 40, )
//
//
//}

func TestResultIter_SelectSorWithWcaIDs(t *testing.T) {
	var v3Db = "root@tcp(127.0.0.1:33306)/cubing_pro?charset=utf8&parseTime=True&loc=Local"
	Db, err := gorm.Open(mysql.New(mysql.Config{DSN: v3Db}), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	r := &ResultIter{
		DB:    Db,
		Cache: cache.New(5*time.Minute, 10*time.Minute),
	}

	out, _ := r.SelectSorWithWcaIDs([]string{
		"2018GUOZ01",
		"2017XUYO01",
		"2018XUEZ01",
		"2018YINZ03",
	}, 1, 10, SelectSorWithWcaIDsOption{
		Events:     []string{"222", "333", "333mbf", "333fm"},
		WithSingle: true,
		WithAvg:    true,
	})
	js, _ := json.MarshalIndent(out, "", "   ")
	fmt.Println(string(js))
}
