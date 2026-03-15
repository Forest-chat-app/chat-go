package router

import (
	v1 "chat-server/api/v1"
	"github.com/gin-gonic/gin"
)

type CdnRouter struct{}

// InitCdnRouter 初始化聊天相关路由
func (s *CdnRouter) InitCdnRouter(apiV1 *gin.RouterGroup) {
	// 聊天相关路由 - 需要认证
	cdnGroup := apiV1.Group("/cdn")
	{
		cdnGroup.GET("/getPostSignature", v1.ApiGroupApp.GetPostSignature)
		cdnGroup.GET("/getCdnUrl", v1.ApiGroupApp.GetCdnUrl)
	}
}
