package router

import (
	v1 "chat-server/api/v1"
	"github.com/gin-gonic/gin"
)

type TokenRouter struct{}

func (s *TokenRouter) InitTokenRouter(apiV1 *gin.RouterGroup) {
	apiV1.Group("/token")
	{
		apiV1.POST("/refreshToken", v1.ApiGroupApp.Register)
	}
}
