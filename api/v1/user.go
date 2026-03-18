package v1

import (
	"chat-server/model/common"
	"chat-server/model/request/user"
	"chat-server/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
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
	data, err := userService.RegisterUser(req, c)
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
		}
		return
	}

	common.Result(c, common.SUCCESS, data)
}

// LoginAccount Login godoc
// @Summary      用户登录
// @Description  用户通过账号密码登录
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request  body      request.LoginRequest  true  "用户登录信息"
// @Success      200      {object}  common.Response
// @Router       /api/v1/user/loginAccount [post]
func (userApi *UserApi) LoginAccount(c *gin.Context) {
	var req user.LoginRequest

	// 校验参数
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}

	// 处理登录业务
	data, err := userService.LoginAccount(req, c)
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
			return
		}
	}

	common.Result(c, common.SUCCESS, data)

}

// GetUserInfo Login godoc
// @Summary      获取用户信息
// @Description  获取用户信息
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        Authorization  header  string  true  "refreshToken"
// @Success      200      {object}  common.Response
// @Router       /api/v1/user/getUserInfo [get]
func (userApi *UserApi) GetUserInfo(c *gin.Context) {
	// 1、获取用户ID
	claims, exists := c.Get("claims")
	if !exists {
		common.Result(c, common.USER_NOT_FOUND)
		return
	}
	userId := claims.(*jwt.Token).Claims.(*utils.AccessToken).UserID

	// 2、获取用户信息
	data, err := userService.GetUserInfo(userId)
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
		}
		return
	}
	common.Result(c, common.SUCCESS, data)
}

// GetUserAndRoom godoc
// @Summary      获取用户和房间信息
// @Description  根据user_id获取用户信息及其所在房间ID列表
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        user_id  query  string  true  "用户ID"
// @Success      200      {object}  common.Response
// @Router       /api/v1/user/getUserAndRoom [get]
func (userApi *UserApi) GetUserAndRoom(c *gin.Context) {
	// 1、校验参数
	var req user.GetUserAndRoomReq
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}

	// 2、执行业务
	data, err := userService.GetUserAndRoom(req.UserId)
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
		}
		return
	}
	common.Result(c, common.SUCCESS, data)
}

// UpdateUserProfile godoc
// @Summary      更新用户信息
// @Description  更新用户信息，空串字段不修改
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request  body      user.UpdateUserProfileRequest  true  "用户信息"
// @Success      200      {object}  common.Response
// @Router       /api/v1/user/updateUserProfile [put]
func (userApi *UserApi) UpdateUserProfile(c *gin.Context) {
	var req user.UpdateUserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}

	claims, exists := c.Get("claims")
	if !exists {
		common.Result(c, common.USER_NOT_FOUND)
		return
	}
	userId := claims.(*jwt.Token).Claims.(*utils.AccessToken).UserID

	data, err := userService.UpdateUserProfile(req, userId)
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
		}
		return
	}
	common.Result(c, common.SUCCESS, data)
}

// UpdateUserAvatar godoc
// @Summary      更新用户头像
// @Description  更新用户头像
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request  body      user.UpdateUserAvatarRequest  true  "用户头像"
// @Success      200      {object}  common.Response
// @Router       /api/v1/user/updateUserAvatar [put]
func (userApi *UserApi) UpdateUserAvatar(c *gin.Context) {
	// 1、绑定请求参数
	var req user.UpdateUserAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}

	// 2、获取用户ID
	claims, exists := c.Get("claims")
	if !exists {
		common.Result(c, common.USER_NOT_FOUND)
		return
	}
	userId := claims.(*jwt.Token).Claims.(*utils.AccessToken).UserID

	// 3、更新用户头像
	err := userService.UpdateUserAvatar(req, userId)
	if err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
		}
		return
	}
	common.Result(c, common.SUCCESS)
}

// UpdateUserPassword godoc
// @Summary      更新用户密码
// @Description  验证旧密码后更新为新密码，成功后自动退出登录
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request  body      user.UpdateUserPasswordRequest  true  "密码信息"
// @Success      200      {object}  common.Response
// @Router       /api/v1/user/updateUserPassword [put]
func (userApi *UserApi) UpdateUserPassword(c *gin.Context) {
	var req user.UpdateUserPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Result(c, common.INVALID_PARAMS)
		return
	}

	if err := userService.UpdateUserPassword(req, c); err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
		}
		return
	}
	common.Result(c, common.SUCCESS)
}

// Logout godoc
// @Summary      退出登录
// @Description  注销用户所有连接并撤销所有token
// @Tags         User
// @Accept       json
// @Produce      json
// @Success      200      {object}  common.Response
// @Router       /api/v1/user/logout [post]
func (userApi *UserApi) Logout(c *gin.Context) {
	if err := userService.Logout(c); err != nil {
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			common.Result(c, serviceErr.GetResponseCode())
		}
		return
	}
	common.Result(c, common.SUCCESS)
}
