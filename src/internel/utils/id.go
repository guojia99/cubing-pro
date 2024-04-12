package utils

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/mozillazg/go-pinyin"
)

// RemoveSpacesAndTranslate 将输入字符串中的空格移除，并将中文字符翻译成拼音后截取前四位英文
func RemoveSpacesAndTranslate(input string) string {
	var result strings.Builder
	input = strings.ReplaceAll(input, " ", "") // 移除空格

	runes := []rune(input)
	for i := 0; i < len(runes); i++ {
		// 判断字符是否是英文、中文、日文、韩文、法文
		if unicode.Is(unicode.Latin, runes[i]) ||
			unicode.Is(unicode.Han, runes[i]) ||
			unicode.Is(unicode.Hiragana, runes[i]) ||
			unicode.Is(unicode.Katakana, runes[i]) ||
			unicode.Is(unicode.Hangul, runes[i]) {
			result.WriteRune(runes[i])
		}
	}
	return result.String()
}
func TranslateEn(input string) string {
	var result strings.Builder
	input = strings.ReplaceAll(input, " ", "") // 移除空格

	runes := []rune(input)
	for i := 0; i < len(runes); i++ {
		// 判断字符是否是英文
		if unicode.Is(unicode.Latin, runes[i]) {
			result.WriteRune(runes[i])
			continue
		}

		// 将中文字符翻译成拼音
		var pinyinStr string
		for _, py := range pinyin.Pinyin(string(runes[i:i+1]), pinyin.NewArgs()) {
			for _, val := range py {
				pinyinStr += val
			}
		}
		result.WriteString(pinyinStr)
	}
	return result.String()
}

// GetIDButNotNumber 生成Cube ID前缀， 2024JIAY 需要自行补充数字
func GetIDButNotNumber(baseName string) string {
	en := TranslateEn(RemoveSpacesAndTranslate(baseName))

	if len(en) >= 4 {
		en = en[:4]
	}

	if len(en) < 4 {
		en += strings.Repeat("0", 4-len(en))
	}

	id := fmt.Sprintf("%d%s", time.Now().Year(), strings.ToUpper(en))

	return id
}
