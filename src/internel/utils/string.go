package utils

import (
	"strings"

	jsoniter "github.com/json-iterator/go"
)

func HideEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email // 如果不是有效的电子邮件地址，返回原始值
	}

	username := parts[0]
	domain := parts[1]

	// 如果用户名长度小于3，则隐藏所有字符；否则，保留前三个字符，其他字符替换为星号
	if len(username) <= 4 {
		username = strings.Repeat("*", len(username))
	} else {
		username = username[:4] + strings.Repeat("*", len(username)-3)
	}

	return username + "@" + domain
}

func ToJSON(in interface{}) string {
	out, _ := jsoniter.MarshalToString(in)
	return out
}
