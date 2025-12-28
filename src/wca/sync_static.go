package wca

import (
	"slices"

	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/log"
)

type staticSyncDone struct {
	Key string `gorm:"primary_key"`
}

func (s *syncer) syncStatics() error {

	_ = s.db.AutoMigrate(&staticSyncDone{})

	var syncFns = map[string]func() error{
		"setStaticPersonRankWithTimer": s.setStaticPersonRankWithTimer,
	}

	var sds []staticSyncDone
	s.db.Find(&sds)

	var syncDoneKey []string
	for _, sd := range sds {
		syncDoneKey = append(syncDoneKey, sd.Key)
	}

	for key, syncFn := range syncFns {
		if slices.Contains(syncDoneKey, key) {
			log.Infof("sync finished for key: %s", key)
			continue
		}
		log.Infof("start sync for key: %s", key)
		if err := syncFn(); err != nil {
			log.Errorf("sync static person rank failed, key: %s, err: %v", key, err)
			continue
		}
		log.Infof("sync finished for key: %s", key)
		st := staticSyncDone{
			Key: key,
		}
		s.db.Save(st)
	}
	return nil
}
