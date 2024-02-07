package utils

import (
	"time"
	"unicode"

	pinyin "github.com/mozillazg/go-pinyin"
)

func IsChineseChar(str string) bool {
	for _, r := range str {
		if unicode.Is(unicode.Scripts["Han"], r) {
			return true
		}
	}
	return false
}

var PinYinArgs = pinyin.NewArgs()

func GetSName(t time.Time, long string) string {
	return ""
}
