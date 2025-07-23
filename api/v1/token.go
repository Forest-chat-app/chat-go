package v1

import (
	"chat-server/global"
	"chat-server/model/common"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
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
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		// 如果 Cookie 不存在或解析失败 (例如客户端没有发送，或者 Cookie 已过期)
		if errors.Is(err, http.ErrNoCookie) {
			global.CHAT_LOG.Warn("RefreshToken: Refresh token cookie not found.")
			common.Result(c, common.REFRESH_TOKEN_INVALID, "Refresh token 解析失败，请重新获取") // 告知客户端需要刷新Token
			return
		}
		// 其他获取 Cookie 的错误
		global.CHAT_LOG.Error("RefreshToken: 获取 refresh token cookie 失败", "err", err)
		common.Result(c, common.ERROR)
		return
	}
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
