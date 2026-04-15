package competition

import (
	"strings"

	"github.com/guojia99/cubing-pro/src/internel/utils"
	jsoniter "github.com/json-iterator/go"
)

// StringListToDB 将多值列表序列化为数据库存储字符串（JSON 数组）
func StringListToDB(parts []string) string {
	var cleaned []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			cleaned = append(cleaned, p)
		}
	}
	cleaned = utils.RemoveDuplicates(cleaned)
	if len(cleaned) == 0 {
		return ""
	}
	return utils.ToJSON(cleaned)
}

// StringListFromDB 从数据库字段解析多值列表，兼容历史单行/逗号分隔数据
func StringListFromDB(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var arr []string
	if err := jsoniter.UnmarshalFromString(s, &arr); err == nil && len(arr) > 0 {
		return arr
	}
	return utils.Split(s, ",")
}
