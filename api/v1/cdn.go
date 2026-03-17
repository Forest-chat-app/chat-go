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

// GetCdnUrl 获取CDN鉴权URL
// @Summary 获取CDN鉴权URL
// @Description 获取CDN鉴权URL
// @Tags 聊天
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.Response "获取CDN鉴权URL成功"
// @Router /api/v1/cdn/getCdnUrl [get]
func (cdnApi *CdnApi) GetCdnUrl(c *gin.Context) {
	// 1、绑定请求参数
	var req cdn.GetCdnUrlRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}

	// 2、处理业务
	data, err := cdnService.GetCdnUrl(req.Url)
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
		}
		return
	}
	common.Result(c, common.SUCCESS, data)

}

// GetUpdateFile 获取更新资源
// @Summary 获取更新资源
// @Description 获取更新资源(rooms/users头像等静态资源)
// @Tags 聊天
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.Response "获取更新资源成功"
// @Router /api/v1/cdn/getUpdateFile [post]
func (cdnApi *CdnApi) GetUpdateFile(c *gin.Context) {
	// 1、绑定请求参数
	var req cdn.GetUpdateFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}

	// 2、处理业务
	data, err := cdnService.GetUpdateFile(req)
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
		}
		return
	}
	common.Result(c, common.SUCCESS, data)
}
