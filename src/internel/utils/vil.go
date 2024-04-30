package utils

import "regexp"

func IsEmailValid(email string) bool {
	// 正则表达式用于验证邮箱格式
	regex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	return regexp.MustCompile(regex).MatchString(email)
}

func IsPhoneNumberValid(phoneNumber string) bool {
	// 正则表达式用于验证手机号格式
	regex := `^1[3456789]\d{9}$`
	return regexp.MustCompile(regex).MatchString(phoneNumber)
}
