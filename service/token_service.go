package service

import (
	"chat-server/constant"
	"chat-server/global"
	"chat-server/middleware"
	"chat-server/model/common"
	"chat-server/utils"
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
)

type TokenService struct{}

// RefreshAccessToken 刷新访问令牌
func (s *TokenService) RefreshAccessToken(refreshTokenString string) (string, error) {
	pipeline := global.CHAT_REDIS.TxPipeline()

	refreshClaims := &middleware.RefreshToken{}
	refreshToken, err := jwt.ParseWithClaims(refreshTokenString, refreshClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(global.CHAT_CONFIG.JWT.Secret), nil
	})

	if err != nil || !refreshToken.Valid {
		return "", common.NewServiceError(common.REFRESH_TOKEN_INVALID)
	}

	// 检查令牌是否被撤销
	isTokenRevoked, err := utils.IsTokenRevoked(refreshClaims.UserID, refreshClaims.TokenID)
	if err != nil {
		return "", common.NewServiceError(common.ERROR)
	}
	if isTokenRevoked {
		return "", common.NewServiceError(common.REFRESH_TOKEN_REVOKED)
	}

	// 删除旧令牌
	ctx := context.Background()
	tokenKey := fmt.Sprintf("%s:%s:%s", constant.RefreshTokenPrefix, refreshClaims.UserID, refreshClaims.TokenID)
	delCmd := pipeline.Del(ctx, tokenKey)
	// 从用户的token集合中移除这个tokenID
	userTokenKey := fmt.Sprintf("%s:%s", constant.UserTokensPrefix, refreshClaims.UserID)
	sremCmd := pipeline.SRem(ctx, userTokenKey, refreshClaims.TokenID)

	// 判断删除操作
	if _, err := pipeline.Exec(ctx); err != nil {
		global.CHAT_LOG.Error("RevokeToken----->删除token失败", "err", err.Error())
		return "", common.NewServiceError(common.ERROR)
	}
	if err := delCmd.Err(); err != nil {
		global.CHAT_LOG.Error("RevokeToken----->删除refresh_token失败", "err", err.Error())
		return "", common.NewServiceError(common.ERROR)
	}
	if err := sremCmd.Err(); err != nil {
		global.CHAT_LOG.Error("RevokeToken----->删除user_tokens失败", "err", err.Error())
		return "", common.NewServiceError(common.ERROR)
	}
	// 获取用户信息
	user, err := utils.GetUserByID(refreshClaims.UserID)
	if err != nil {
		return "", err
	}

	// 生成新的令牌对
	tokenPair, err := utils.GenerateTokenPair(user.ID, user.UserAccount)
	if err != nil {
		return "", err
	}

	return tokenPair.AccessToken, nil
}
