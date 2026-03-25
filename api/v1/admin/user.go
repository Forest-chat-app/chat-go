package admin

import (
	"chat-server/model/common"
	reqAdmin "chat-server/model/request/admin"
	adminSvc "chat-server/service/admin"
	"errors"

	"github.com/gin-gonic/gin"
)

type AdminUserApi struct{}

// ListUsers 用户列表
func (a *AdminUserApi) ListUsers(c *gin.Context) {
	var req reqAdmin.ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	data, err := adminSvc.AdminUserServiceApp.ListUsers(req)
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

// GetUserDetail 用户详情
func (a *AdminUserApi) GetUserDetail(c *gin.Context) {
	var req reqAdmin.GetUserDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	data, err := adminSvc.AdminUserServiceApp.GetUserDetail(req.UserId)
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

// UpdateUserProfile 修改用户信息（不含头像）
func (a *AdminUserApi) UpdateUserProfile(c *gin.Context) {
	var req reqAdmin.UpdateAdminUserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	if err := adminSvc.AdminUserServiceApp.UpdateUserProfile(req); err != nil {
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

// UpdateUserAvatar 修改用户头像
func (a *AdminUserApi) UpdateUserAvatar(c *gin.Context) {
	var req reqAdmin.UpdateAdminUserAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	if err := adminSvc.AdminUserServiceApp.UpdateUserAvatar(req); err != nil {
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

// ResetUserPassword 重置用户密码
func (a *AdminUserApi) ResetUserPassword(c *gin.Context) {
	var req reqAdmin.ResetUserPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	if err := adminSvc.AdminUserServiceApp.ResetUserPassword(req); err != nil {
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

// UnbanUser 解封用户
func (a *AdminUserApi) UnbanUser(c *gin.Context) {
	var req reqAdmin.UnbanUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	if err := adminSvc.AdminUserServiceApp.UnbanUser(req); err != nil {
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
func (a *AdminUserApi) BanUser(c *gin.Context) {
	var req reqAdmin.BanUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	if err := adminSvc.AdminUserServiceApp.BanUser(req); err != nil {
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

// AdminCreateUser 管理端创建用户
func (a *AdminUserApi) AdminCreateUser(c *gin.Context) {
	var req reqAdmin.AdminCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	if err := adminSvc.AdminUserServiceApp.AdminCreateUser(req); err != nil {
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
