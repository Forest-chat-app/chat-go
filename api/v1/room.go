package v1

import (
	"chat-server/utils"
	"errors"

	//"chat-server/middleware"
	"chat-server/model/common"
	"chat-server/model/request/room"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
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
	claims, exists := c.Get("claims")
	if !exists {
		common.Result(c, common.USER_NOT_FOUND)
		return
	}
	userId := claims.(*jwt.Token).Claims.(*utils.AccessToken).UserID

	// 处理业务
	data, err := roomService.CreateRoom(req, userId)
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
		}
		return
	}
	common.Result(c, common.SUCCESS, data)
}

// SearchRoom 搜索聊天室
// @Summary 搜索聊天室
// @Description 搜索聊天室
// @Tags 聊天
// @Accept json
// @Produce json
// @Param request body room.searchRoom true "搜索聊天室请求"
// @Security BearerAuth
// @Success      200      {object}  common.Response
// @Router /api/v1/room/searchRoom [get]
func (r *RoomApi) SearchRoom(c *gin.Context) {
	// 1、校验参数
	req := room.SearchRoomRequest{}
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}

	// 2、处理业务
	data, err := roomService.SearchRoom(req)
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
		}
		return
	}
	common.Result(c, common.SUCCESS, data)
}

// HotRoom 获取热门聊天室
// @Summary 获取热门聊天室
// @Description 获取热门聊天室
// @Tags 聊天
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success      200      {object}  common.Response
// @Router /api/v1/room/hotRoom [get]
func (r *RoomApi) HotRoom(c *gin.Context) {
	// 1、处理业务
	data, err := roomService.HotRoom()
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
		}
		return
	}
	common.Result(c, common.SUCCESS, data)
}

// JoinRoom 加入聊天室
// @Summary 加入聊天室
// @Description 加入聊天室
// @Tags 聊天
// @Accept json
// @Produce json
// @Param request body room.joinRoom true "加入聊天室请求"
// @Security BearerAuth
// @Success      200      {object}  common.Response
// @Router /api/v1/room/joinRoom [post]
func (r *RoomApi) JoinRoom(c *gin.Context) {
	// 1、校验参数
	req := room.JoinRoomRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}

	// 2、获取userId
	claims, exists := c.Get("claims")
	if !exists {
		common.Result(c, common.USER_NOT_FOUND)
		return
	}
	userId := claims.(*jwt.Token).Claims.(*utils.AccessToken).UserID

	// 3、处理业务
	_, err := roomService.JoinRoom(req, userId)
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
		}
		return
	}
	common.Result(c, common.SUCCESS)
}

// QuitRoom 退出聊天室
// @Summary 退出聊天室
// @Description 退出聊天室
// @Tags 聊天
// @Accept json
// @Produce json
// @Param request body room.quitRoom true "退出聊天室请求"
// @Security BearerAuth
// @Success      200      {object}  common.Response
// @Router /api/v1/room/quitRoom [delete]
func (r *RoomApi) QuitRoom(c *gin.Context) {
	// 1、校验参数
	req := room.QuitRoomRequest{}
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}

	// 2、获取userId
	claims, exists := c.Get("claims")
	if !exists {
		common.Result(c, common.USER_NOT_FOUND)
		return
	}
	userId := claims.(*jwt.Token).Claims.(*utils.AccessToken).UserID

	// 3、退出房间
	data, err := roomService.QuitRoom(req, userId)
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
		}
		return
	}
	common.Result(c, common.SUCCESS, data)
}

// PreviewRoom 预览聊天室
// @Summary 预览聊天室
// @Description 预览聊天室
// @Tags 聊天
// @Accept json
// @Produce json
// @Param request body room.previewRoom true "预览聊天室请求"
// @Security BearerAuth
// @Success      200      {object}  common.Response
// @Router /api/v1/room/previewRoom [get]
func (r *RoomApi) PreviewRoom(c *gin.Context) {
	// 1、校验参数
	req := room.PreviewRoomRequest{}
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}

	// 2、获取userId
	claims, exists := c.Get("claims")
	if !exists {
		common.Result(c, common.USER_NOT_FOUND)
		return
	}
	userId := claims.(*jwt.Token).Claims.(*utils.AccessToken).UserID

	// 3、加入房间预览
	roomService.PreviewRoom(req, userId)
	common.Result(c, common.SUCCESS)

}

// DeleteRoomMembers 移除房间成员
// @Summary 移除房间成员
// @Description 移除房间成员（仅群主/群管理员可操作）
// @Tags 聊天
// @Accept json
// @Produce json
// @Param request body room.DeleteRoomMembersRequest true "移除房间成员请求"
// @Security BearerAuth
// @Success      200      {object}  common.Response
// @Router /api/v1/room/deleteRoomMembers [delete]
func (r *RoomApi) DeleteRoomMembers(c *gin.Context) {
	// 1、校验参数
	req := room.DeleteRoomMembersRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}

	// 2、获取userId
	claims, exists := c.Get("claims")
	if !exists {
		common.Result(c, common.USER_NOT_FOUND)
		return
	}
	userId := claims.(*jwt.Token).Claims.(*utils.AccessToken).UserID

	// 3、处理业务
	if err := roomService.DeleteRoomMembers(req, userId); err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
		}
		return
	}
	common.Result(c, common.SUCCESS)
}

// GetRoomMembers 获取房间成员列表
// @Summary 获取房间成员列表
// @Description 获取房间成员列表
// @Tags 聊天
// @Accept json
// @Produce json
// @Param room_id query string true "房间ID"
// @Security BearerAuth
// @Success      200      {object}  common.Response
// @Router /api/v1/room/getRoomMembers [get]
func (r *RoomApi) GetRoomMembers(c *gin.Context) {
	// 1、校验参数
	req := room.GetRoomMembersRequest{}
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}

	// 2、处理业务
	data, err := roomService.GetRoomMembers(req)
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
		}
		return
	}
	common.Result(c, common.SUCCESS, data)
}

// LeavePreview 预览聊天室
// @Summary 退出预览聊天室
// @Description 预览聊天室
// @Tags 聊天
// @Accept json
// @Produce json
// @Param request body room.leavePreview true "预览聊天室请求"
// @Security BearerAuth
// @Success      200      {object}  common.Response
// @Router /api/v1/room/leavePreview [delete]
func (r *RoomApi) LeavePreview(c *gin.Context) {
	// 1、校验参数
	req := room.LeavePreviewRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}

	// 2、获取userId
	claims, exists := c.Get("claims")
	if !exists {
		common.Result(c, common.USER_NOT_FOUND)
		return
	}
	userId := claims.(*jwt.Token).Claims.(*utils.AccessToken).UserID

	// 3、加入房间预览
	roomService.LeavePreview(req, userId)
	common.Result(c, common.SUCCESS)
}

// UpdateRoomAvatar godoc
// @Summary      更新房间头像
// @Description  更新房间头像
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request  body      user.UpdateRoomAvatarRequest  true  "房间头像"
// @Success      200      {object}  common.Response
// @Router       /api/v1/user/updateRoomAvatar [put]
func (r *RoomApi) UpdateRoomAvatar(c *gin.Context) {
	// 1、绑定请求参数
	var req room.UpdateRoomAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}

	// 3、更新用户头像
	err := roomService.UpdateRoomAvatar(req)
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
		}
		return
	}
	common.Result(c, common.SUCCESS)
}
