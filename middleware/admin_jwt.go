package middleware

import (
	"chat-server/constant"
	"chat-server/global"
	"chat-server/model/common"
	"chat-server/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// AdminJWTAuth 管理端JWT认证中间件
func AdminJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, common.Response{
				Code: common.ADMIN_UNAUTHORIZED.Code,
				Msg:  common.ADMIN_UNAUTHORIZED.Msg,
				Data: map[string]interface{}{},
			})
			c.Abort()
			return
		}

		token = strings.TrimPrefix(token, "Bearer ")
		claims, err := jwt.ParseWithClaims(token, &utils.AdminTokenClaims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(global.CHAT_CONFIG.JWT.Secret), nil
		})
		if err != nil || !claims.Valid {
			c.JSON(http.StatusUnauthorized, common.Response{
				Code: common.ADMIN_UNAUTHORIZED.Code,
				Msg:  common.ADMIN_UNAUTHORIZED.Msg,
				Data: map[string]interface{}{},
			})
			c.Abort()
			return
		}

		adminClaims := claims.Claims.(*utils.AdminTokenClaims)
		// 验证必须是管理员角色
		if adminClaims.RoleId != constant.RoleSuperAdmin && adminClaims.RoleId != constant.RoleCommonAdmin {
			c.JSON(http.StatusForbidden, common.Response{
				Code: common.ADMIN_FORBIDDEN.Code,
				Msg:  common.ADMIN_FORBIDDEN.Msg,
				Data: map[string]interface{}{},
			})
			c.Abort()
			return
		}

		c.Set("admin_claims", adminClaims)
		c.Next()
	}
}

// SuperAdminOnly 仅超级管理员可访问的中间件（需配合 AdminJWTAuth 使用）
func SuperAdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("admin_claims")
		if !exists {
			c.JSON(http.StatusForbidden, common.Response{
				Code: common.ADMIN_FORBIDDEN.Code,
				Msg:  common.ADMIN_FORBIDDEN.Msg,
				Data: map[string]interface{}{},
			})
			c.Abort()
			return
		}
		adminClaims := claims.(*utils.AdminTokenClaims)
		if adminClaims.RoleId != constant.RoleSuperAdmin {
			c.JSON(http.StatusForbidden, common.Response{
				Code: common.ADMIN_ONLY_SUPER.Code,
				Msg:  common.ADMIN_ONLY_SUPER.Msg,
				Data: map[string]interface{}{},
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
