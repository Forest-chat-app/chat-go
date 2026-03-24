package admin

import (
	"chat-server/model/common"
	reqAdmin "chat-server/model/request/admin"
	adminSvc "chat-server/service/admin"
	"errors"

	"github.com/gin-gonic/gin"
)

type AdminRoomApi struct{}

// ListRooms 房间列表
func (a *AdminRoomApi) ListRooms(c *gin.Context) {
	var req reqAdmin.ListRoomsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	data, err := adminSvc.AdminRoomServiceApp.ListRooms(req)
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
			return
		}
		common.Result(c, common.ERROR)
		return
	}
	common.Result(c, common.SUCCESS, data)
}

// GetRoomDetail 房间详情
func (a *AdminRoomApi) GetRoomDetail(c *gin.Context) {
	var req reqAdmin.GetRoomDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	data, err := adminSvc.AdminRoomServiceApp.GetRoomDetail(req.RoomId)
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
			return
		}
		common.Result(c, common.ERROR)
		return
	}
	common.Result(c, common.SUCCESS, data)
}

// CreateRoom 创建房间
func (a *AdminRoomApi) CreateRoom(c *gin.Context) {
	var req reqAdmin.AdminCreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	data, err := adminSvc.AdminRoomServiceApp.CreateRoom(req)
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
			return
		}
		common.Result(c, common.ERROR)
		return
	}
	common.Result(c, common.SUCCESS, data)
}

// UpdateRoomInfo 修改房间信息（不含头像）
func (a *AdminRoomApi) UpdateRoomInfo(c *gin.Context) {
	var req reqAdmin.AdminUpdateRoomInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	if err := adminSvc.AdminRoomServiceApp.UpdateRoomInfo(req); err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
			return
		}
		common.Result(c, common.ERROR)
		return
	}
	common.Result(c, common.SUCCESS)
}

// UpdateRoomAvatar 修改房间头像
func (a *AdminRoomApi) UpdateRoomAvatar(c *gin.Context) {
	var req reqAdmin.AdminUpdateRoomAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	if err := adminSvc.AdminRoomServiceApp.UpdateRoomAvatar(req); err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
			return
		}
		common.Result(c, common.ERROR)
		return
	}
	common.Result(c, common.SUCCESS)
}

// DeleteRoom 解散房间
func (a *AdminRoomApi) DeleteRoom(c *gin.Context) {
	var req reqAdmin.AdminDeleteRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	if err := adminSvc.AdminRoomServiceApp.DeleteRoom(req); err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
			return
		}
		common.Result(c, common.ERROR)
		return
	}
	common.Result(c, common.SUCCESS)
}

// GetRoomMembers 获取房间成员列表
func (a *AdminRoomApi) GetRoomMembers(c *gin.Context) {
	var req reqAdmin.AdminGetRoomMembersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	data, err := adminSvc.AdminRoomServiceApp.GetRoomMembers(req)
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
			return
		}
		common.Result(c, common.ERROR)
		return
	}
	common.Result(c, common.SUCCESS, data)
}

// RemoveRoomMember 移除房间成员
func (a *AdminRoomApi) RemoveRoomMember(c *gin.Context) {
	var req reqAdmin.AdminRemoveRoomMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	if err := adminSvc.AdminRoomServiceApp.RemoveRoomMember(req); err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
			return
		}
		common.Result(c, common.ERROR)
		return
	}
	common.Result(c, common.SUCCESS)
}
