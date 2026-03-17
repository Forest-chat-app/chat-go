package service

import (
	"chat-server/constant"
	"chat-server/global"
	"chat-server/model/common"
	"chat-server/model/mysql"
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

// GetUpdateFile 获取需要更新的资源（头像等）
func (cdnService *CdnService) GetUpdateFile(req cdn.GetUpdateFileRequest) (map[string]interface{}, error) {
	tx := global.CHAT_MYSQL

	// --- 处理 rooms ---
	roomResults := make([]cdn.GetUpdateFileRoomResult, 0)
	if len(req.Rooms) > 0 {
		// 构建 room_id -> 客户端 updated_at 的映射
		roomIdToClientTs := make(map[string]int64, len(req.Rooms))
		roomIds := make([]string, 0, len(req.Rooms))
		for _, r := range req.Rooms {
			roomIds = append(roomIds, r.RoomId)
			roomIdToClientTs[r.RoomId] = r.UpdatedAt
		}

		// 一次查询取出所有匹配 room 的 id、avatar、updated_at
		type roomRow struct {
			ID        string
			Avatar    string
			UpdatedAt int64
		}
		var rows []roomRow
		tx.Model(&mysql.Room{}).Select("id, avatar, updated_at").
			Where("id IN ?", roomIds).Scan(&rows)

		// Go 侧过滤：只保留 db.updated_at > client.updated_at 的记录
		for _, row := range rows {
			if row.UpdatedAt > roomIdToClientTs[row.ID] {
				roomResults = append(roomResults, cdn.GetUpdateFileRoomResult{
					RoomId: row.ID,
					Avatar: utils.GenerateCdnUrl(row.Avatar),
				})
			}
		}
	}

	// --- 处理 users ---
	userResults := make([]cdn.GetUpdateFileUserResult, 0)
	if len(req.Users) > 0 {
		userIdToClientTs := make(map[string]int64, len(req.Users))
		userIds := make([]string, 0, len(req.Users))
		for _, u := range req.Users {
			userIds = append(userIds, u.UserId)
			userIdToClientTs[u.UserId] = u.UpdatedAt
		}

		type userRow struct {
			ID        string
			Avatar    string
			UpdatedAt int64
		}
		var uRows []userRow
		tx.Model(&mysql.User{}).Select("id, avatar, updated_at").
			Where("id IN ?", userIds).Scan(&uRows)

		for _, row := range uRows {
			if row.UpdatedAt > userIdToClientTs[row.ID] {
				userResults = append(userResults, cdn.GetUpdateFileUserResult{
					UserId: row.ID,
					Avatar: utils.GenerateCdnUrl(row.Avatar),
				})
			}
		}
	}

	return map[string]interface{}{
		"rooms": roomResults,
		"users": userResults,
	}, nil
}
