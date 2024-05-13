package initer

import (
	"log"

	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"gorm.io/gorm"
)

var events = append(wcaEvents, append(otherEvents, notCubes...)...)

func initEvent(db *gorm.DB) error {
	for _, ev := range events {
		var evDb event.Event
		if err := db.First(&evDb, "id = ?", ev.ID).Error; err == nil {
			continue
		}

		if err := db.Create(&ev).Error; err != nil {
			log.Default().Println(err)
		}
	}
	return nil
}
