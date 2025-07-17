package v1

import (
	"chat-server/model/common"
	"chat-server/model/request/user"
	"errors"
	"github.com/gin-gonic/gin"
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

	common.Result(c, common.SUCCESS, tokenPair)

}
