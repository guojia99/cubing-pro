package crawler

import (
	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
)

type SendEmail struct {
	basemodel.Model

	Email string // 防重复发送的
	Type  string // 类型 cubing, wca
	Key   string // 唯一识别ID
}
