package router

import (
	v1 "chat-server/api/v1"
	"github.com/gin-gonic/gin"
)

type TokenRouter struct{}

func (s *TokenRouter) InitTokenRouter(apiV1 *gin.RouterGroup) {
	tokenGroup := apiV1.Group("/token")
	{
		tokenGroup.POST("/refreshToken", v1.ApiGroupApp.RefreshToken)
	}
}
