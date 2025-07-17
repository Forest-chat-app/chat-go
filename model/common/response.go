package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Result 返回统一格式的响应
// 如果不提供data参数，则默认为空map
func Result(c *gin.Context, respCode ResponseCode, data ...interface{}) {
	// 设置默认的空数据
	var responseData interface{} = map[string]interface{}{}

	// 如果提供了data参数，则使用它
	if len(data) > 0 {
		responseData = data[0]
	}

	c.JSON(http.StatusOK, Response{
		Code: respCode.Code,
		Msg:  respCode.Msg,
		Data: responseData,
	})
}
