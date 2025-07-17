package utils

import (
	"regexp"
)

// VerifyEmail 验证邮箱格式
func VerifyEmail(email string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(email)
}
