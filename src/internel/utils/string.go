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

func ReplaceAll(s, new string, old ...string) string {
	for _, o := range old {
		s = strings.ReplaceAll(s, o, new)
	}
	return s
}

func RemoveEmptyLines(input string) string {
	// 按照换行符分割字符串，得到每一行
	lines := strings.Split(input, "\n")

	// 创建一个新的切片，用于存储非空行
	var nonEmptyLines []string

	// 遍历每一行，去掉完全为空的行
	for _, line := range lines {
		// TrimSpace 去除每一行的前后空白字符（包括空格和换行符）
		if strings.TrimSpace(line) != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}

	// 将剩下的行重新组合成一个字符串，使用换行符分隔
	return strings.Join(nonEmptyLines, "\n")
}
