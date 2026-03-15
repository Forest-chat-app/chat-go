package service

import (
	"chat-server/constant"
	"chat-server/global"
	"chat-server/model/common"
	"chat-server/model/request/cdn"
	"chat-server/utils"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/aliyun/credentials-go/credentials"
	"hash"
	"time"
)

type CdnService struct{}

// GetPostSignature 获取上传OSS凭证
func (cdnService *CdnService) GetPostSignature(subDir string) (cdn.GetPostSignatureResponse, error) {
	cfg := global.CHAT_CONFIG.CdnConfig
	product := "oss"
	if !constant.AllowedDirs[subDir] {
		return cdn.GetPostSignatureResponse{}, common.NewServiceError(common.NOT_ALLOWED_DIR)
	}
	targetDir := subDir + "/"

	// 1. 初始化凭证（STS模式）
	credConfig := new(credentials.Config).
		SetType("ram_role_arn").
		SetAccessKeyId(cfg.AccessKeyId).
		SetAccessKeySecret(cfg.AccessKeySecret).
		SetRoleArn(cfg.RoleArn).
		SetRoleSessionName(cfg.SessionName).
		SetRoleSessionExpiration(cfg.TokenExpire)

	provider, _ := credentials.NewCredential(credConfig)
	cred, err := provider.GetCredential()
	if err != nil {
		global.CHAT_LOG.Error("GetPostSignature-->获取凭证失败", "err", err)
		return cdn.GetPostSignatureResponse{}, common.NewServiceError(common.ERROR)
	}

	// 2. 构建 Policy
	now := time.Now().UTC()
	date := now.Format("20060102")
	expiration := now.Add(time.Duration(cfg.TokenExpire) * time.Second)

	policyMap := map[string]any{
		"expiration": expiration.Format("2006-01-02T15:04:05.000Z"),
		"conditions": []any{
			map[string]string{"bucket": cfg.BucketName},
			[]string{"starts-with", "$key", targetDir}, // 必须以此目录开头
			[]any{"content-length-range", 0, 104857600},
			map[string]string{"x-oss-signature-version": "OSS4-HMAC-SHA256"},
			map[string]string{"x-oss-credential": fmt.Sprintf("%v/%v/%v/%v/aliyun_v4_request", *cred.AccessKeyId, date, cfg.Region, product)},
			map[string]string{"x-oss-date": now.Format("20060102T150405Z")},
			map[string]string{"x-oss-security-token": *cred.SecurityToken},
		},
	}

	policyBytes, _ := json.Marshal(policyMap)
	stringToSign := base64.StdEncoding.EncodeToString(policyBytes)

	// 3. V4 签名计算 (HMAC-SHA256)
	hmacHash := func() hash.Hash { return sha256.New() }
	signingKey := "aliyun_v4" + *cred.AccessKeySecret

	h1 := hmac.New(hmacHash, []byte(signingKey))
	h1.Write([]byte(date))
	h2 := hmac.New(hmacHash, h1.Sum(nil))
	h2.Write([]byte(cfg.Region))
	h3 := hmac.New(hmacHash, h2.Sum(nil))
	h3.Write([]byte(product))
	h4 := hmac.New(hmacHash, h3.Sum(nil))
	h4.Write([]byte("aliyun_v4_request"))

	h := hmac.New(hmacHash, h4.Sum(nil))
	h.Write([]byte(stringToSign))
	signature := hex.EncodeToString(h.Sum(nil))

	return cdn.GetPostSignatureResponse{
		Policy:           stringToSign,
		SecurityToken:    *cred.SecurityToken,
		SignatureVersion: "OSS4-HMAC-SHA256",
		Credential:       fmt.Sprintf("%v/%v/%v/%v/aliyun_v4_request", *cred.AccessKeyId, date, cfg.Region, product),
		Date:             now.Format("20060102T150405Z"),
		Signature:        signature,
		Host:             cfg.OssHost,
		Dir:              targetDir,
	}, nil
}

// GetCdnUrl 获取CDN鉴权URL
func (cdnService *CdnService) GetCdnUrl(url string) (cdn.GetCdnUrlResponse, error) {
	return cdn.GetCdnUrlResponse{
		Url: utils.GenerateCdnUrl(url),
	}, nil
}
