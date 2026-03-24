package admin

import (
	"chat-server/model/common"
	reqAdmin "chat-server/model/request/admin"
	adminSvc "chat-server/service/admin"
	"chat-server/utils"
	"errors"

	"github.com/gin-gonic/gin"
)

type AdminAuthApi struct{}

// Login 管理员登录
func (a *AdminAuthApi) Login(c *gin.Context) {
	var req reqAdmin.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}

	data, err := adminSvc.AdminAuthServiceApp.Login(req)
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

// GetSelf 获取当前管理员信息
func (a *AdminAuthApi) GetSelf(c *gin.Context) {
	claims := c.MustGet("admin_claims").(*utils.AdminTokenClaims)
	data, err := adminSvc.AdminAuthServiceApp.GetSelf(claims)
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
