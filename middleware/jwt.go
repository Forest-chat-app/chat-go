package middleware

import (
	"chat-server/global"
	"chat-server/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// 不需要验证的路径
var excludePaths = []string{
	"/api/v1/user/register",
	"/api/v1/user/loginAccount",
	"/api/v1/user/test",
	"/token/refreshToken",
	"/swagger/",
	"/api/v1/admin/login",
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1、检查是否在排除路径列表中
		path := c.Request.URL.Path
		for _, excludePath := range excludePaths {
			if strings.HasPrefix(path, excludePath) {
				// 如果在排除列表中，跳过验证
				c.Next()
				return
			}
		}

		// 2、不在排除列表中，执行JWT验证
		// 2.1 获取token
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			token = c.Query("token")
		}
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"data":    nil,
				"message": "未授权",
			})
			c.Abort()
			return
		}

		// 2.2执行验证
		token = strings.TrimPrefix(token, "Bearer ")
		claims, err := jwt.ParseWithClaims(token, &utils.AccessToken{}, func(token *jwt.Token) (interface{}, error) {
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

		// 3、将用户信息存储到上下文中
		c.Set("claims", claims)
		c.Next()
	}
}
