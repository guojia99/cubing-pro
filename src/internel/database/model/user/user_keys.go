package user

import basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"

type UserKVType = int

const (
	UserKVTypeString UserKVType = iota + 1
	UserKVTypeInt
	UserKVTypeJSON
	UserKVTypeYaml
	UserKVTypeMD
)

// UserKV 用户kv数据
type UserKV struct {
	basemodel.DBModel

	UserId uint   `gorm:"index:,unique,composite:AssUserKV"`
	Key    string `gorm:"index:,unique,composite:AssUserKV"`

	Value string     `gorm:"column:value"`
	Type  UserKVType `gorm:"column:type"`
}
