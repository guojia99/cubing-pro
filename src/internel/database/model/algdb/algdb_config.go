package algdb

import (
	"encoding/json"

	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	"gorm.io/gorm"
)

type AlgConfig struct {
	basemodel.StringIDModel

	Value       string `gorm:"column:value"` // json
	Description string `gorm:"column:description"`
}

func GetKeyJSONValue(db *gorm.DB, key string, value interface{}) error {
	var kv AlgConfig
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
	var kv = AlgConfig{
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
