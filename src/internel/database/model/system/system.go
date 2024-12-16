package system

import (
	"encoding/json"
	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	"gorm.io/gorm"
)

// KeyValue 保存一些系统kv数据的数据库
type KeyValue struct {
	basemodel.StringIDModel

	Value       string `gorm:"column:value"`
	Description string `gorm:"column:description"`
}

func GetKeyJSONValue(db *gorm.DB, key string, value interface{}) error {
	var kv KeyValue
	if err := db.First(&kv, "id = ?", key).Error; err != nil {
		return err
	}
	return json.Unmarshal([]byte(kv.Value), &value)
}

func SetKeyJSONValue(db *gorm.DB, key string, value interface{}, description string) error {
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
	return db.Save(&kv).Error
}
