package router

import (
	v1 "chat-server/api/v1"

	"github.com/gin-gonic/gin"
)

type ChatRouter struct{}

// InitChatRouter 初始化聊天相关路由
func (s *ChatRouter) InitChatRouter(apiV1 *gin.RouterGroup) {
	// 聊天相关路由 - 需要认证
	chatGroup := apiV1.Group("/chat")
	{
		chatGroup.GET("/webSocketHandler", v1.ApiGroupApp.WebSocketHandler)
		chatGroup.GET("/getHistoryMsg", v1.ApiGroupApp.GetHistoryMsg)
		chatGroup.GET("/searchChat", v1.ApiGroupApp.SearchChat)
	}
}
