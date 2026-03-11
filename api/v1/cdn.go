package v1

import (
	"chat-server/model/common"
	"chat-server/model/request/cdn"
	"errors"
	"github.com/gin-gonic/gin"
)

type CdnApi struct{}

// GetPostSignature 获取OSS上传签名
// @Summary 获取OSS上传签名
// @Description 获取OSS上传签名
// @Tags 聊天
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.Response "获取OSS上传签名成功"
// @Router /api/v1/cdn/getPostSignature [get]
func (cdnApi *CdnApi) GetPostSignature(c *gin.Context) {
	// 1、校验参数
	var req = cdn.GetPostSignatureRequest{}
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}

	// 2、处理业务
	data, err := cdnService.GetPostSignature(req.SubDir)
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
		}
		return
	}
	common.Result(c, common.SUCCESS, data)
}
