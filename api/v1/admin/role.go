package admin

import (
	"chat-server/model/common"
	reqAdmin "chat-server/model/request/admin"
	adminSvc "chat-server/service/admin"
	"chat-server/utils"
	"errors"

	"github.com/gin-gonic/gin"
)

type AdminRoleApi struct{}

// ListRoles 角色列表
func (a *AdminRoleApi) ListRoles(c *gin.Context) {
	var req reqAdmin.ListRolesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	data, err := adminSvc.AdminRoleServiceApp.ListRoles(req)
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

// UpdateRole 修改角色描述
func (a *AdminRoleApi) UpdateRole(c *gin.Context) {
	var req reqAdmin.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	if err := adminSvc.AdminRoleServiceApp.UpdateRole(req); err != nil {
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

// ListAdmins 管理员列表（超级管理员专用）
func (a *AdminRoleApi) ListAdmins(c *gin.Context) {
	var req reqAdmin.AdminListAdminsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	data, err := adminSvc.AdminRoleServiceApp.ListAdmins(req)
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

// CreateAdmin 提升为管理员（超级管理员专用）
func (a *AdminRoleApi) CreateAdmin(c *gin.Context) {
	var req reqAdmin.AdminCreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	if err := adminSvc.AdminRoleServiceApp.CreateAdmin(req); err != nil {
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

// DeleteAdmin 撤销管理员（超级管理员专用）
func (a *AdminRoleApi) DeleteAdmin(c *gin.Context) {
	var req reqAdmin.AdminDeleteAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	claims := c.MustGet("admin_claims").(*utils.AdminTokenClaims)
	if err := adminSvc.AdminRoleServiceApp.DeleteAdmin(req, claims.UserID); err != nil {
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
