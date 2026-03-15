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
		roomGroup.POST("/createRoom", v1.ApiGroupApp.CreateRoom)
		roomGroup.GET("/searchRoom", v1.ApiGroupApp.SearchRoom)
		roomGroup.GET("/hotRoom", v1.ApiGroupApp.HotRoom)
		roomGroup.POST("/joinRoom", v1.ApiGroupApp.JoinRoom)
		roomGroup.DELETE("/quitRoom", v1.ApiGroupApp.QuitRoom)
		roomGroup.GET("/previewRoom", v1.ApiGroupApp.PreviewRoom)
		roomGroup.DELETE("/deleteRoomMembers", v1.ApiGroupApp.DeleteRoomMembers)
		roomGroup.GET("/getRoomMembers", v1.ApiGroupApp.GetRoomMembers)
		roomGroup.DELETE("/leavePreview", v1.ApiGroupApp.LeavePreview)
	}

}
