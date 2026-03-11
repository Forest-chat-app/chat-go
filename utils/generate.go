package utils

import (
	"chat-server/global"
	"crypto/md5"
	"fmt"
	"github.com/google/uuid"
	"time"
)

// GenerateUUid 生成uuid
func GenerateUUid() string {
	return uuid.New().String()
}

// GenerateCdnUrl 生成CDN访问URL
func GenerateCdnUrl(uri string) string {
	// 1. 准备基础参数
	domain := global.CHAT_CONFIG.CdnConfig.Domain
	privateKey := global.CHAT_CONFIG.CdnConfig.PrivateKey
	expireDuration := global.CHAT_CONFIG.CdnConfig.ExpireDuration
	timestamp := time.Now().Unix() + expireDuration
	rand := "0"
	uid := "0"

	// 2. 构造待加密字符串: /Filename-timestamp-rand-uid-PrivateKey
	// 注意：uri 必须以 / 开头
	originStr := fmt.Sprintf("%s-%d-%s-%s-%s", uri, timestamp, rand, uid, privateKey)

	// 3. 计算 MD5
	hash := md5.Sum([]byte(originStr))
	md5hash := fmt.Sprintf("%x", hash)

	// 4. 拼接最终 URL
	return fmt.Sprintf("%s%s?auth_key=%d-%s-%s-%s", domain, uri, timestamp, rand, uid, md5hash)
}
