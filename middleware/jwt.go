package middleware

import (
	"chat-server/global"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
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
	UserID  string `json:"user_id"`
	TokenID string `json:"token_id"`
	jwt.RegisteredClaims
}

// 不需要验证的路径
var excludePaths = []string{
	"/api/v1/user/register",
	"/api/v1/user/login",
	"/api/v1/user/test",
	"/token/refreshToken",
	"/swagger/",
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前请求路径
		path := c.Request.URL.Path

		// 检查是否在排除路径列表中
		for _, excludePath := range excludePaths {
			if strings.HasPrefix(path, excludePath) {
				// 如果在排除列表中，跳过验证
				c.Next()
				return
			}
		}

		// 不在排除列表中，执行JWT验证
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"data":    nil,
				"message": "未授权",
			})
			c.Abort()
			return
		}

		// 移除Bearer前缀
		if strings.HasPrefix(token, "Bearer ") {
			token = token[7:]
		}

		// 验证token
		claims, err := jwt.ParseWithClaims(token, &AccessToken{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(global.CHAT_CONFIG.JWT.Secret), nil
		})
		if err != nil || !claims.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"data":    nil,
				"message": "未授权",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("claims", claims)
		c.Next()
	}
}
