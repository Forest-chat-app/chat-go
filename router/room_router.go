package router

import (
	v1 "chat-server/api/v1"
	"github.com/gin-gonic/gin"
)

type RoomRouter struct{}

func (r *RoomRouter) InitRoomRouter(apiV1 *gin.RouterGroup) {
	// 聊天室相关路由 - 需要认证
	roomGroup := apiV1.Group("/room")
	{
		roomGroup.GET("/createRoom", v1.ApiGroupApp.WebSocketHandler)
	}

}
