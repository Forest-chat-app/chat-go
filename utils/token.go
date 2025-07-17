package utils

import (
	"chat-server/constant"
	"chat-server/global"
	"chat-server/middleware"
	"chat-server/model"
	"chat-server/model/common"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// GenerateTokenPair 生成JWT令牌
func GenerateTokenPair(userID string, userAccount string) (*middleware.TokenPair, error) {
	// 生成唯一的令牌ID
	tokenID := uuid.New().String()

	// 创建访问令牌
	accessClaims := middleware.AccessToken{
		UserID:      userID,
		UserAccount: userAccount,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * time.Duration(global.CHAT_CONFIG.JWT.AccessTime))), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),                                                                     // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),                                                                     // 生效时间
			Issuer:    global.CHAT_CONFIG.JWT.Issuer,                                                                      // 签发人
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(global.CHAT_CONFIG.JWT.Secret))
	if err != nil {
		global.CHAT_LOG.Error("GenerateTokenPair-->签名生成accessToken失败", "err", err)
		return nil, common.NewServiceError(common.ERROR)
	}

	// 创建刷新令牌
	refreshClaims := middleware.RefreshToken{
		UserID:  userID,
		TokenID: tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * time.Duration(global.CHAT_CONFIG.JWT.RefreshTime))), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),                                                                         // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),                                                                         // 生效时间
			Issuer:    global.CHAT_CONFIG.JWT.Issuer,                                                                          // 签发人
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(global.CHAT_CONFIG.JWT.Secret))
	if err != nil {
		global.CHAT_LOG.Error("GenerateTokenPair-->签名生成refreshToken失败", "err", err)
		return nil, common.NewServiceError(common.ERROR)
	}

	return &middleware.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    global.CHAT_CONFIG.JWT.AccessTime,
	}, nil
}

// ParseAccessToken 解析访问令牌
func ParseAccessToken(tokenString string) (*middleware.AccessToken, error) {
	claims := &middleware.AccessToken{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(global.CHAT_CONFIG.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// StoreRefreshToken 存储刷新令牌到数据库
func StoreRefreshToken(userID string, tokenID string, platform string) error {
	// RefreshToken存储配置
	ctx := context.Background()
	tokenKey := fmt.Sprintf("%s:%s:%s", constant.RefreshTokenPrefix, userID, tokenID)
	tokenData := map[string]interface{}{
		"userId":     userID,
		"tokenId":    tokenID,
		"expiresAt":  UnixToUTCMillisTimestamp(time.Now().Add(time.Hour * 24 * time.Duration(global.CHAT_CONFIG.JWT.RefreshTime)).Unix()),
		"created_at": GetUTCMillisTimestamp(),
		"platform":   platform,
	}
	tokenJson, err := json.Marshal(tokenData)
	if err != nil {
		global.CHAT_LOG.Error("存储RefreshToken序列化失败", "err", err.Error())
		return err
	}

	pipeline := global.CHAT_REDIS.TxPipeline()
	userTokenKey := fmt.Sprintf("%s:%s", constant.UserTokensPrefix, userID)
	// 1、RefreshToken存入redis 2、把RefreshToken的tokenId存储进用户令牌集合
	pipeline.Set(ctx, tokenKey, tokenJson, time.Hour*24*time.Duration(global.CHAT_CONFIG.JWT.RefreshTime))
	pipeline.SAdd(ctx, userTokenKey, tokenID)
	pipeline.Expire(ctx, userTokenKey, time.Hour*24*time.Duration(global.CHAT_CONFIG.JWT.UserTokensTime))
	if _, err := pipeline.Exec(ctx); err != nil {
		global.CHAT_LOG.Error("存储RefreshToken失败", "err", err.Error())
		return err
	}

	return nil
}

// IsTokenRevoked 检查令牌是否被撤销
func IsTokenRevoked(userID string, tokenID string) (bool, error) {
	redis := global.CHAT_REDIS
	ctx := context.Background()

	// 实现检查逻辑
	tokenKey := fmt.Sprintf("%s:%s:%s", constant.RefreshTokenPrefix, userID, tokenID)
	exists, err := redis.Exists(ctx, tokenKey).Result()
	if err != nil {
		global.CHAT_LOG.Error("检查RefreshToken是否被撤销，操作失败", "err", err.Error())
		return false, err
	}
	if exists == 0 {
		return true, nil
	}

	return false, nil
}

// GetUserByID 根据ID获取用户
func GetUserByID(userID string) (*model.User, error) {
	mysql := global.CHAT_MYSQL

	// 实现用户查询逻辑
	queryUser := model.User{}
	err := mysql.Where("id = ?", userID).First(&queryUser).Error
	if err != nil {
		return nil, common.NewServiceError(common.USER_ID_NOT_FOUND)
	}

	return &queryUser, nil
}

// RevokeAllUserTokens 撤销用户的所有令牌（登出所有设备）
func RevokeAllUserTokens(userID uint) error {
	// 实现撤销逻辑
	// 例如：将用户所有令牌的isRevoked设为true
	return nil
}

// RevokeToken 撤销特定令牌（单设备登出）
func RevokeToken(userID string, tokenID string) error {
	pipeline := global.CHAT_REDIS.TxPipeline()
	ctx := context.Background()

	// 1、把refresh_token删除
	tokenKey := fmt.Sprintf("%s:%s:%s", constant.RefreshTokenPrefix, userID, tokenID)
	delCmd := pipeline.Del(ctx, tokenKey)
	// 2、从user_tokens集合中删除
	userTokenKey := fmt.Sprintf("%s:%s", constant.UserTokensPrefix, userID)
	sremCmd := pipeline.SRem(ctx, userTokenKey, tokenID)

	// 判断操作
	if _, err := pipeline.Exec(ctx); err != nil {
		global.CHAT_LOG.Error("RevokeToken----->删除token失败", "err", err.Error())
		return err
	}
	if err := delCmd.Err(); err != nil {
		global.CHAT_LOG.Error("RevokeToken----->删除refresh_token失败", "err", err.Error())
		return err
	}
	if err := sremCmd.Err(); err != nil {
		global.CHAT_LOG.Error("RevokeToken----->删除user_tokens失败", "err", err.Error())
		return err
	}

	return nil
}
