package v1

import (
	//"chat-server/middleware"
	"chat-server/model/common"
	"chat-server/model/request/room"
	"github.com/gin-gonic/gin"
	//"github.com/golang-jwt/jwt/v4"
)

type RoomApi struct{}

// CreateRoom 创建聊天室
// @Summary 创建聊天室
// @Description 创建聊天室
// @Tags 聊天
// @Accept json
// @Produce json
// @Param request body room.CreateRoom true "创建聊天室请求"
// @Security BearerAuth
// @Success      200      {object}  common.Response
// @Router /api/v1/room/createRoom [post]
func (r *RoomApi) CreateRoom(c *gin.Context) {
	// 校验参数
	req := room.CreateRoomRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	// 获取userId
	//claims, exists := c.Get("claims")
	//if !exists {
	//	common.Result(c, common.USER_NOT_FOUND)
	//	return
	//}
	//userId := claims.(*jwt.Token).Claims.(*middleware.AccessToken).UserID

	// 处理业务

}
