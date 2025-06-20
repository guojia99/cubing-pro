package system

import (
	"encoding/json"
	"sync"
	"time"

	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	"gorm.io/gorm"
)

var systemKeyLock = sync.Mutex{}

// KeyValue 保存一些系统kv数据的数据库
type KeyValue struct {
	basemodel.StringIDModel

	Value       string `gorm:"column:value"`
	Description string `gorm:"column:description"`
}

func GetKeyJSONValue(db *gorm.DB, key string, value interface{}) error {
	systemKeyLock.Lock()
	defer systemKeyLock.Unlock()

	var kv KeyValue
	if err := db.First(&kv, "id = ?", key).Error; err != nil {
		return err
	}
	return json.Unmarshal([]byte(kv.Value), &value)
}

func SetKeyJSONValue(db *gorm.DB, key string, value interface{}, description string) error {
	systemKeyLock.Lock()
	defer systemKeyLock.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var kv = KeyValue{
		StringIDModel: basemodel.StringIDModel{
			ID: key,
		},
	}
	db.First(&kv, "id = ?", key)
	if description != "" {
		kv.Description = description
	}

	kv.Value = string(data)

	// 避免0值
	now := time.Now()
	if kv.CreatedAt.IsZero() {
		kv.CreatedAt = now
	}
	kv.UpdatedAt = now

	return db.Save(&kv).Error
}
