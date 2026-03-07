package service

import (
	"chat-server/constant"
	"chat-server/global"
	"chat-server/model/common"
	"chat-server/utils"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type TokenService struct{}

// RefreshAccessToken 刷新访问令牌
func (s *TokenService) RefreshAccessToken(refreshTokenString string, c *gin.Context) (map[string]interface{}, error) {
	pipeline := global.CHAT_REDIS.TxPipeline()

	// 1、解析refresh_token
	refreshClaims := &utils.RefreshToken{}
	refreshToken, err := jwt.ParseWithClaims(refreshTokenString, refreshClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(global.CHAT_CONFIG.JWT.Secret), nil
	})
	if err != nil || !refreshToken.Valid {
		return nil, common.NewServiceError(common.REFRESH_TOKEN_INVALID)
	}

	// 2、检查令牌是否被撤销
	isTokenExist, err := utils.IsTokenExist(refreshClaims.UserID, refreshClaims.Platform)
	if err != nil {
		return nil, common.NewServiceError(common.ERROR)
	}
	if !isTokenExist {
		return nil, common.NewServiceError(common.REFRESH_TOKEN_REVOKED)
	}

	// 3、删除旧令牌
	ctx := context.Background()
	tokenKey := fmt.Sprintf("%s:%s:%s", constant.RefreshTokenPrefix, refreshClaims.UserID, refreshClaims.Platform)
	delCmd := pipeline.Del(ctx, tokenKey)
	if _, err := pipeline.Exec(ctx); err != nil {
		global.CHAT_LOG.Error("RevokeToken----->删除token失败", "err", err.Error())
		return nil, common.NewServiceError(common.ERROR)
	}
	if err := delCmd.Err(); err != nil {
		global.CHAT_LOG.Error("RevokeToken----->删除refresh_token失败", "err", err.Error())
		return nil, common.NewServiceError(common.ERROR)
	}

	// 4、生成新的令牌对
	user, err := utils.GetUserByID(refreshClaims.UserID)
	if err != nil {
		return nil, err
	}
	tokenPair, err := utils.GenerateTokenPair(user.ID, user.UserAccount, refreshClaims.Platform)
	if err != nil {
		return nil, err
	}
	// 4.1 在redis保存RefreshToken状态
	err = utils.StoreRefreshToken(user.ID, tokenPair.RefreshToken, refreshClaims.Platform)
	if err != nil {
		global.CHAT_LOG.Error("LoginAccount-->保存RefreshToken状态失败", "err", err)
		return nil, common.NewServiceError(common.ERROR)
	}

	// 5、设置cookie
	cookieName := "refresh_token"
	maxAge := int(time.Duration(global.CHAT_CONFIG.JWT.RefreshTime) * 24 * time.Hour / time.Second)
	c.SetCookie(
		cookieName,             // Cookie 名称
		tokenPair.RefreshToken, // Cookie 值
		maxAge,                 // Cookie 的最大生命周期（秒）
		"/",                    // Cookie 路径，"/" 表示所有路径都可访问
		"",                     // Cookie 作用域，生产环境应替换为你的域名，例如 "api.yourdomain.com" 或 "yourdomain.com"
		// 开发测试时可以用 "localhost" 或留空
		false, // Secure: 只在 HTTPS 连接中发送此 Cookie
		true,  // HttpOnly: 无法通过 JavaScript 访问此 Cookie
	)

	// 6、返回数据
	data := map[string]interface{}{
		"access_token": tokenPair.AccessToken,
		"expires_in":   tokenPair.ExpiresIn,
	}

	return data, nil
}
