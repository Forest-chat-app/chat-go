package v1

import "github.com/gin-gonic/gin"

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

}
