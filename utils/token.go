package utils

import (
	"chat-server/constant"
	"chat-server/global"
	"chat-server/model/common"
	"chat-server/model/mysql"
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type AccessToken struct {
	UserID      string `json:"user_id"`
	UserAccount string `json:"user_account"`
	jwt.RegisteredClaims
}

type RefreshToken struct {
	UserID   string `json:"user_id"`
	Platform string `json:"platform"`
	jwt.RegisteredClaims
}

// GenerateTokenPair 生成JWT令牌
func GenerateTokenPair(userID string, userAccount string, platform string) (*TokenPair, error) {
	// 创建访问令牌
	accessClaims := AccessToken{
		UserID:      userID,
		UserAccount: userAccount,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Minute * time.Duration(global.CHAT_CONFIG.JWT.AccessTime))), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),                                                                     // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now().UTC()),                                                                     // 生效时间
			Issuer:    global.CHAT_CONFIG.JWT.Issuer,                                                                            // 签发人
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(global.CHAT_CONFIG.JWT.Secret))
	if err != nil {
		global.CHAT_LOG.Error("GenerateTokenPair-->签名生成accessToken失败", "err", err)
		return nil, common.NewServiceError(common.ERROR)
	}

	// 创建刷新令牌
	refreshClaims := RefreshToken{
		UserID:   userID,
		Platform: platform,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour * 24 * time.Duration(global.CHAT_CONFIG.JWT.RefreshTime))), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),                                                                         // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now().UTC()),                                                                         // 生效时间
			Issuer:    global.CHAT_CONFIG.JWT.Issuer,                                                                                // 签发人
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(global.CHAT_CONFIG.JWT.Secret))
	if err != nil {
		global.CHAT_LOG.Error("GenerateTokenPair-->签名生成refreshToken失败", "err", err)
		return nil, common.NewServiceError(common.ERROR)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    global.CHAT_CONFIG.JWT.AccessTime,
	}, nil
}

// StoreRefreshToken 存储刷新令牌到数据库
func StoreRefreshToken(userID string, refreshToken string, platform string) error {
	// 1、获取redis句柄
	pipeline := global.CHAT_REDIS.TxPipeline()
	ctx := context.Background()

	// 2、设置key-value
	tokenKey := fmt.Sprintf("%s:%s:%s", constant.RefreshTokenPrefix, userID, platform)
	tokenValue := refreshToken

	// 3、RefreshToken存入redis
	pipeline.Set(ctx, tokenKey, tokenValue, time.Hour*24*time.Duration(global.CHAT_CONFIG.JWT.RefreshTime))
	if _, err := pipeline.Exec(ctx); err != nil {
		global.CHAT_LOG.Error("存储RefreshToken失败", "err", err.Error())
		return err
	}
	return nil
}

// IsTokenExist 检查令牌是否存在
func IsTokenExist(userID string, platform string) (bool, error) {
	// 1、获取句柄
	redis := global.CHAT_REDIS
	ctx := context.Background()

	// 2、检查是否存在
	tokenKey := fmt.Sprintf("%s:%s:%s", constant.RefreshTokenPrefix, userID, platform)
	exists, err := redis.Exists(ctx, tokenKey).Result()
	if err != nil {
		global.CHAT_LOG.Error("检查RefreshToken是否被撤销，操作失败", "err", err.Error())
		return false, err
	}
	if exists == 0 {
		return false, nil
	}

	return true, nil
}

// GetUserByID 根据ID获取用户
func GetUserByID(userID string) (*mysql.User, error) {
	tx := global.CHAT_MYSQL

	// 实现用户查询逻辑
	queryUser := mysql.User{}
	err := tx.Where("id = ?", userID).First(&queryUser).Error
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
func RevokeToken(userID string, platform string) error {
	pipeline := global.CHAT_REDIS.TxPipeline()
	ctx := context.Background()

	// 1、把refresh_token删除
	tokenKey := fmt.Sprintf("%s:%s:%s", constant.RefreshTokenPrefix, userID, platform)
	delCmd := pipeline.Del(ctx, tokenKey)

	// 1.1 判断操作
	if _, err := pipeline.Exec(ctx); err != nil {
		global.CHAT_LOG.Error("RevokeToken----->删除token失败", "err", err.Error())
		return err
	}
	if err := delCmd.Err(); err != nil {
		global.CHAT_LOG.Error("RevokeToken----->删除refresh_token失败", "err", err.Error())
		return err
	}

	return nil
}
