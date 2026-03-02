package utils

import (
	"regexp"
)

// VerifyEmail 验证邮箱格式
func VerifyEmail(email string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(email)
}

// VerifyAvatar 验证头像url格式
func VerifyAvatar(avatar string) bool {
	return regexp.MustCompile(`^https?:\/\/[^\s]+$`).MatchString(avatar)
}

// VerifyString 验证字符串是否非空
func VerifyString(str string) bool {
	return str != ""
}
