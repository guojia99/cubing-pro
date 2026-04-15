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
	basemodel.Model

	UserId uint   `json:"userId" gorm:"uniqueIndex:idx_user_key"`       // 联合唯一索引名：idx_user_key
	Key    string `json:"key" gorm:"uniqueIndex:idx_user_key;size:191"` // MySQL 索引需指定长度，避免 longtext

	Value string     `json:"value" gorm:"column:value"`
	Type  UserKVType `json:"type" gorm:"column:type"`
}

const MaxKVLength = 1024 * 1024 * 2 // 2MB

var WhitelistKeys = []string{
	"blind_tightening_assistant", // 盲拧助手
	"algorithm_config",           // 公式库配置
	"website_ui_config",          // 网页 UI 配置（推荐）
	"websize_ui_config",          // 历史拼写，兼容旧数据
	"group_timer_ui_config",      // 计时器配置
}

func IsInWhitelist(key string) bool {
	for _, v := range WhitelistKeys {
		if v == key {
			return true
		}
	}
	return false
}
