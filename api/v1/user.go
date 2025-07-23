package v1

import (
	"chat-server/global"
	"chat-server/model/common"
	"chat-server/model/request/user"
	"errors"
	"github.com/gin-gonic/gin"
	"time"
)

type UserApi struct{}

// Test godoc
// @Summary      测试接口
// @Description  测试接口，返回一些示例数据
// @Tags         User
// @Accept       json
// @Produce      json
// @Success      200  {object}  common.Response
// @Router       /user/test [get]
func (userApi *UserApi) Test(c *gin.Context) {
	var xxx = []int{1, 2, 3}
	var yyy = map[string]interface{}{
		"name": "张三",
		"age":  18,
	}
	var zzz = map[string]interface{}{
		"xxx": xxx,
		"yyy": yyy,
	}
	common.Result(c, common.SUCCESS, zzz)
}

// Register godoc
// @Summary      用户注册
// @Description  创建新用户账号
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request  body      request.RegisterRequest  true  "用户注册信息"
// @Success      200      {object}  common.Response
// @Router       /api/v1/user/register [post]
func (userApi *UserApi) Register(c *gin.Context) {
	var req user.RegisterRequest

	// 校验参数
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}

	//处理注册业务
	tokenPair, err := userService.RegisterUser(req.UserAccount, req.Password, req.Email, req.Platform)
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
		}
		return
	}

	// 设置refresh_token为http only
	refreshToken := tokenPair.RefreshToken
	cookieName := "refresh_token"
	maxAge := int(time.Duration(global.CHAT_CONFIG.JWT.RefreshTime) * 24 * time.Hour / time.Second)
	// 设置 Cookie
	c.SetCookie(
		cookieName,   // Cookie 名称
		refreshToken, // Cookie 值
		maxAge,       // Cookie 的最大生命周期（秒）
		"/",          // Cookie 路径，"/" 表示所有路径都可访问
		"",           // Cookie 作用域，生产环境应替换为你的域名，例如 "api.yourdomain.com" 或 "yourdomain.com"
		// 开发测试时可以用 "localhost" 或留空
		false, // Secure: 只在 HTTPS 连接中发送此 Cookie
		true,  // HttpOnly: 无法通过 JavaScript 访问此 Cookie
	)
	tokenPair.RefreshToken = ""

	common.Result(c, common.SUCCESS, tokenPair)
}

// LoginAccount Login godoc
// @Summary      用户登录
// @Description  用户通过账号密码登录
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request  body      request.LoginRequest  true  "用户登录信息"
// @Success      200      {object}  common.Response
// @Router       /api/v1/user/login [post]
func (userApi *UserApi) LoginAccount(c *gin.Context) {
	var req user.LoginRequest

	// 校验参数
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}

	// 处理登录业务
	tokenPair, err := userService.LoginAccount(req.UserAccount, req.Password, req.Platform)
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
			return
		}
	}

	// 设置refresh_token为http only
	refreshToken := tokenPair.RefreshToken
	cookieName := "refresh_token"
	maxAge := int(time.Duration(global.CHAT_CONFIG.JWT.RefreshTime) * 24 * time.Hour / time.Second)
	// 设置 Cookie
	c.SetCookie(
		cookieName,   // Cookie 名称
		refreshToken, // Cookie 值
		maxAge,       // Cookie 的最大生命周期（秒）
		"/",          // Cookie 路径，"/" 表示所有路径都可访问
		"",           // Cookie 作用域，生产环境应替换为你的域名，例如 "api.yourdomain.com" 或 "yourdomain.com"
		// 开发测试时可以用 "localhost" 或留空
		false, // Secure: 只在 HTTPS 连接中发送此 Cookie
		true,  // HttpOnly: 无法通过 JavaScript 访问此 Cookie
	)

	tokenPair.RefreshToken = ""

	common.Result(c, common.SUCCESS, tokenPair)

}
