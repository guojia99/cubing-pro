package system

import (
	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
)

// KeyValue 保存一些系统kv数据的数据库
type KeyValue struct {
	basemodel.StringIDModel

	Value       string `gorm:"column:value"`
	Description string `gorm:"column:description"`
}
