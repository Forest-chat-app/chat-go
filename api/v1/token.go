package v1

import (
	"chat-server/model/common"
	"errors"
	"github.com/gin-gonic/gin"
)

type TokenApi struct{}

// RefreshToken godoc
// @Summary      刷新令牌
// @Description  通过刷新令牌，刷新refreshToken和accessToken
// @Tags         Token
// @Accept       json
// @Produce      json
// @Param        Authorization  header  string  true  "refreshToken"
// @Success      200  {object}  common.Response
// @Router       /api/v1/token/refreshToken [get]
func (a *TokenApi) RefreshToken(c *gin.Context) {
	refreshToken := c.GetHeader("Authorization")

	// 生成新令牌对
	tokenPair, err := tokenService.RefreshAccessToken(refreshToken)
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
		}
	}
	common.Result(c, common.SUCCESS, tokenPair)
}
