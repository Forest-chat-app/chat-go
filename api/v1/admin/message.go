package admin

import (
	"chat-server/model/common"
	reqAdmin "chat-server/model/request/admin"
	adminSvc "chat-server/service/admin"
	"errors"

	"github.com/gin-gonic/gin"
)

type AdminMessageApi struct{}

// ListMessages 房间消息列表
func (a *AdminMessageApi) ListMessages(c *gin.Context) {
	var req reqAdmin.AdminListMessagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	data, err := adminSvc.AdminMessageServiceApp.ListMessages(req)
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

// SearchMessages 全局搜索消息
func (a *AdminMessageApi) SearchMessages(c *gin.Context) {
	var req reqAdmin.AdminSearchMessagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	data, err := adminSvc.AdminMessageServiceApp.SearchMessages(req)
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

// DeleteMessage 删除消息
func (a *AdminMessageApi) DeleteMessage(c *gin.Context) {
	var req reqAdmin.AdminDeleteMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	if err := adminSvc.AdminMessageServiceApp.DeleteMessage(req); err != nil {
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
