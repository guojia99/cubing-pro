package crawler

import (
	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
)

type SendEmail struct {
	basemodel.DBModel

	Email string
	Type  string
	Key   string // 唯一识别ID
}
