package service

import (
	"chat-server/constant"
	"chat-server/core"
	"chat-server/global"
	"chat-server/model/common"
	"chat-server/model/mysql"
	"chat-server/model/request/user"
	"chat-server/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"time"
)

type UserService struct{}

// RegisterUser 注册用户
func (s *UserService) RegisterUser(req user.RegisterRequest, c *gin.Context) (map[string]interface{}, error) {
	// 1、开启Mysql事务
	tx := global.CHAT_MYSQL.Begin()
	if tx.Error != nil {
		global.CHAT_LOG.Error("RegisterUser-->开启Mysql事务失败", "err", tx.Error.Error())
		return nil, common.NewServiceError(common.ERROR)
	}
	defer func() {
		if r := recover(); r != nil {
			global.CHAT_LOG.Error("RegisterUser-->捕捉到panic", "err", r)
			tx.Rollback()
		} else if tx.Error != nil {
			global.CHAT_LOG.Error("RegisterUser-->捕捉到tx.Error", "err", r)
			tx.Rollback()
		} else {
			tx.Commit()
			global.CHAT_LOG.Info(fmt.Sprintf("RegisterUser-->%s-->mysql无报错", req.UserAccount))
		}
	}()

	// 2、检查用户名是否已存在
	var count int64
	err := tx.Model(&mysql.User{}).Where("user_account = ?", req.UserAccount).Count(&count).Error
	if err != nil {
		tx.Error = err
		global.CHAT_LOG.Error("RegisterUser-->检查用户账号，数据库操作错误", "err", err)
		return nil, common.NewServiceError(common.ERROR)
	}
	if count > 0 {
		return nil, common.NewServiceError(common.USER_ACCOUNT_EXISTS)
	}

	// 3、校验参数
	if len(req.Password) <= 0 {
		return nil, common.NewServiceError(common.PASSWORD_INVALID)
	}
	hashedPassword, err := utils.GenerateFromPassword(req.Password)
	if err != nil || hashedPassword == "" {
		global.CHAT_LOG.Error("RegisterUser-->加密密码出错", "err", err)
		return nil, common.NewServiceError(common.ERROR)
	}

	// 4、创建新用户
	userID := utils.GenerateUUid()
	createUser := mysql.User{
		ID:          userID,
		UserAccount: req.UserAccount,
		Password:    hashedPassword,
		Nickname:    constant.Nickname,
		Email:       constant.Email,
		Avatar:      constant.Avatar,
		CreatedAt:   utils.GetUTCMillisTimestamp(),
		UpdatedAt:   utils.GetUTCMillisTimestamp(),
	}
	err = tx.Create(&createUser).Error
	if err != nil {
		tx.Error = err
		global.CHAT_LOG.Error("RegisterUser-->创建用户，数据库操作错误", "err", err)
		return nil, common.NewServiceError(common.ERROR)
	}
	// 4.1、创建用户角色
	var queryRole mysql.Role
	tx.First(&queryRole, "role_id = ?", constant.RoleNormalUser)
	if queryRole.ID == "" {
		return nil, common.NewServiceError(common.ROLE_NOT_FOUND)
	}
	userRole := mysql.UserRole{
		ID:        utils.GenerateUUid(),
		UserID:    userID,
		RoleID:    queryRole.ID,
		CreatedAt: utils.GetUTCMillisTimestamp(),
		UpdatedAt: utils.GetUTCMillisTimestamp(),
	}
	err = tx.Create(&userRole).Error
	if err != nil {
		tx.Error = err
		global.CHAT_LOG.Error("RegisterUser-->创建用户角色，数据库操作错误", "err", err)
		return nil, common.NewServiceError(common.ERROR)
	}

	// 5、创建用户成功, 生成token
	tokenPair, err := utils.GenerateTokenPair(userID, req.UserAccount, req.Platform)
	if err != nil {
		tx.Error = err
		global.CHAT_LOG.Error("RegisterUser-->生成token失败", "err", err)
		return nil, common.NewServiceError(common.GENERATE_TOKEN_ERROR)
	}
	// 5.1、在redis保存RefreshToken状态
	err = utils.StoreRefreshToken(userID, tokenPair.RefreshToken, req.Platform)
	if err != nil {
		tx.Error = err
		global.CHAT_LOG.Error("RegisterUser-->保存RefreshToken状态失败", "err", err)
		return nil, common.NewServiceError(common.ERROR)
	}

	// 6、处理返回数据
	// 6.1设置refresh_token为http only
	cookieName := "refresh_token"
	maxAge := int(time.Duration(global.CHAT_CONFIG.JWT.RefreshTime) * 24 * time.Hour / time.Second)
	c.SetCookie(
		cookieName,             // Cookie 名称
		tokenPair.RefreshToken, // Cookie 值
		maxAge,                 // Cookie 的最大生命周期（秒）
		"/",                    // Cookie 路径，"/" 表示所有路径都可访问
		"",                     // Cookie 作用域，生产环境应替换为你的域名，例如 "api.yourdomain.com" 或 "yourdomain.com"
		// 开发测试时可以用 "localhost" 或留空
		false, // Secure: 只在 HTTPS 连接中发送此 Cookie
		true,  // HttpOnly: 无法通过 JavaScript 访问此 Cookie
	)
	// 7、返回数据
	data := map[string]interface{}{
		"access_token": tokenPair.AccessToken,
		"expires_in":   tokenPair.ExpiresIn,
	}

	return data, nil
}

// LoginAccount 账号登录
func (s *UserService) LoginAccount(req user.LoginRequest, c *gin.Context) (map[string]interface{}, error) {
	tx := global.CHAT_MYSQL.Begin()

	// 1、开启mysql事务
	if tx.Error != nil {
		global.CHAT_LOG.Error("LoginAccount-->开启Mysql事务失败", "err", tx.Error.Error())
		return nil, common.NewServiceError(common.ERROR)
	}
	defer func() {
		if r := recover(); r != nil {
			global.CHAT_LOG.Error("LoginAccount-->捕捉到panic", "err", r)
			tx.Rollback()
		} else if tx.Error != nil {
			global.CHAT_LOG.Error("LoginAccount-->捕捉到tx.Error", "err", r)
			tx.Rollback()
		} else {
			tx.Commit()
			global.CHAT_LOG.Info(fmt.Sprintf("LoginAccount-->%s-->mysql无报错", req.UserAccount))
		}
	}()

	// 2、验证user
	var queryUser mysql.User
	err := tx.Where("user_account = ?", req.UserAccount).First(&queryUser).Error
	if err != nil {
		global.CHAT_LOG.Error("LoginAccount-->检查用户账号，数据库操作错误", "err", err)
		return nil, common.NewServiceError(common.USER_ACCOUNT_NOT_FOUND)
	}
	// 2.1 验证密码
	match, err := utils.CompareHashAndPassword(queryUser.Password, req.Password)
	if err != nil {
		return nil, common.NewServiceError(common.ERROR)
	}
	if !match {
		return nil, common.NewServiceError(common.PASSWORD_INVALID)
	}

	// 2.2 检查redis是否已存在该登录平台的token，只允许单平台登录
	isTokenExist, err := utils.IsTokenExist(queryUser.ID, req.Platform)
	if err != nil {
		return nil, common.NewServiceError(common.ERROR)
	}
	if isTokenExist {
		return nil, common.NewServiceError(common.PLATFORM_LOGGED_IN)
	}

	// 3、通过验证，下发token
	tokenPair, err := utils.GenerateTokenPair(queryUser.ID, queryUser.UserAccount, req.Platform)
	if err != nil {
		global.CHAT_LOG.Error("LoginAccount-->生成token失败", "err", err)
		return nil, common.NewServiceError(common.GENERATE_TOKEN_ERROR)
	}

	// 3.1 在redis保存RefreshToken状态
	err = utils.StoreRefreshToken(queryUser.ID, tokenPair.RefreshToken, req.Platform)
	if err != nil {
		global.CHAT_LOG.Error("LoginAccount-->保存RefreshToken状态失败", "err", err)
		return nil, common.NewServiceError(common.ERROR)
	}

	// 3.2、设置refresh_token为http only
	cookieName := "refresh_token"
	maxAge := int(time.Duration(global.CHAT_CONFIG.JWT.RefreshTime) * 24 * time.Hour / time.Second)
	c.SetCookie(
		cookieName,             // Cookie 名称
		tokenPair.RefreshToken, // Cookie 值
		maxAge,                 // Cookie 的最大生命周期（秒）
		"/",                    // Cookie 路径，"/" 表示所有路径都可访问
		"",                     // Cookie 作用域，生产环境应替换为你的域名，例如 "api.yourdomain.com" 或 "yourdomain.com"
		// 开发测试时可以用 "localhost" 或留空
		false, // Secure: 只在 HTTPS 连接中发送此 Cookie
		true,  // HttpOnly: 无法通过 JavaScript 访问此 Cookie
	)

	// 4、返回数据
	data := map[string]interface{}{
		"user":         queryUser,
		"access_token": tokenPair.AccessToken,
		"expires_in":   tokenPair.ExpiresIn,
	}

	return data, nil
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(userId string) (map[string]interface{}, error) {
	tx := global.CHAT_MYSQL
	// 1、获取用户信息
	user, err := utils.GetUserByID(userId)
	if err != nil {
		global.CHAT_LOG.Error("GetUserInfo-->获取用户信息失败", "err", err)
		var serviceErr common.ServiceErr
		if errors.As(err, &serviceErr) {
			return nil, err
		}
		return nil, err
	}
	// 2、获取用户角色信息
	var userRole mysql.UserRole
	tx.First(&userRole, "user_id = ?", userId)
	if userRole.ID == "" {
		return nil, common.NewServiceError(common.USER_ROLE_NOT_FOUND)
	}
	var role mysql.Role
	tx.First(&role, "id = ?", userRole.RoleID)
	if role.ID == "" {
		return nil, common.NewServiceError(common.ROLE_NOT_FOUND)
	}
	// 3、获取该用户所有房间
	var roomIDs []string
	tx.Model(&mysql.RoomMembers{}).Where("user_id = ?", userId).Pluck("room_id", &roomIDs)
	var rooms []mysql.Room
	if len(roomIDs) > 0 {
		tx.Where("id IN ?", roomIDs).Find(&rooms)
	}
	// 4、返回数据
	data := map[string]interface{}{
		"user":  user,
		"role":  role,
		"rooms": rooms,
	}
	return data, nil
}

// UpdateUserInfo 更新用户信息
func (s *UserService) UpdateUserInfo(req user.UpdateUserInfoRequest, userId string) (map[string]interface{}, error) {
	// 1、开启mysql事务
	tx := global.CHAT_MYSQL.Begin()
	if tx.Error != nil {
		global.CHAT_LOG.Error("UpdateUserInfo-->开启Mysql事务失败", "err", tx.Error.Error())
		return nil, common.NewServiceError(common.ERROR)
	}
	defer func() {
		if r := recover(); r != nil {
			global.CHAT_LOG.Error("UpdateUserInfo-->捕捉到panic", "err", r)
			tx.Rollback()
		} else if tx.Error != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// 2、构建需要更新的字段
	updates := map[string]interface{}{}
	if req.UserAccount != "" {
		var count int64
		tx.Model(&mysql.User{}).Where("user_account = ? AND id != ?", req.UserAccount, userId).Count(&count)
		if count > 0 {
			return nil, common.NewServiceError(common.USER_ACCOUNT_DUPLICATE)
		}
		updates["user_account"] = req.UserAccount
	}
	if req.Email != "" {
		var count int64
		tx.Model(&mysql.User{}).Where("email = ? AND id != ?", req.Email, userId).Count(&count)
		if count > 0 {
			return nil, common.NewServiceError(common.EMAIL_DUPLICATE)
		}
		updates["email"] = req.Email
	}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if len(updates) == 0 {
		return nil, nil
	}
	updates["updated_at"] = utils.GetUTCMillisTimestamp()

	// 3、更新信息
	if err := tx.Model(&mysql.User{}).Where("id = ?", userId).Updates(updates).Error; err != nil {
		tx.Error = err
		global.CHAT_LOG.Error("UpdateUserInfo-->更新用户信息失败", "err", err)
		return nil, common.NewServiceError(common.ERROR)
	}
	return nil, nil
}

// UpdateUserPassword 更新用户密码
func (s *UserService) UpdateUserPassword(req user.UpdateUserPasswordRequest, c *gin.Context) error {
	// 1、检查refresh_token
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		// 如果 Cookie 不存在或解析失败 (例如客户端没有发送，或者 Cookie 已过期)
		if errors.Is(err, http.ErrNoCookie) {
			global.CHAT_LOG.Warn("RefreshToken: Refresh token cookie not found.")
			return common.NewServiceError(common.REFRESH_TOKEN_INVALID)
		}
		// 其他获取 Cookie 的错误
		global.CHAT_LOG.Error("RefreshToken: 获取 refresh token cookie 失败", "err", err)
		return common.NewServiceError(common.ERROR)
	}
	refreshClaims := &utils.RefreshToken{}
	_, err = jwt.ParseWithClaims(refreshToken, refreshClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(global.CHAT_CONFIG.JWT.Secret), nil
	})

	// 1.1 检查令牌是否被撤销
	isTokenExist, err := utils.IsTokenExist(refreshClaims.UserID, refreshClaims.Platform)
	if err != nil {
		return common.NewServiceError(common.ERROR)
	}
	if !isTokenExist {
		return common.NewServiceError(common.REFRESH_TOKEN_REVOKED)
	}

	// 2、获取user比对密码
	queryUser, err := utils.GetUserByID(refreshClaims.UserID)
	if err != nil {
		return err
	}
	match, err := utils.CompareHashAndPassword(queryUser.Password, req.OldPassword)
	if err != nil {
		global.CHAT_LOG.Error("UpdateUserPassword-->比对密码出错", "err", err)
		return common.NewServiceError(common.ERROR)
	}
	if !match {
		return common.NewServiceError(common.PASSWORD_INVALID)
	}
	// 2.1 加密新密码
	hashedPassword, err := utils.GenerateFromPassword(req.NewPassword)
	if err != nil || hashedPassword == "" {
		global.CHAT_LOG.Error("UpdateUserPassword-->加密新密码出错", "err", err)
		return common.NewServiceError(common.ERROR)
	}
	// 2.2 更新密码
	if err := global.CHAT_MYSQL.Model(&mysql.User{}).Where("id = ?", queryUser.ID).
		Updates(map[string]interface{}{
			"password":   hashedPassword,
			"updated_at": utils.GetUTCMillisTimestamp(),
		}).Error; err != nil {
		global.CHAT_LOG.Error("UpdateUserPassword-->更新密码失败", "err", err)
		return common.NewServiceError(common.ERROR)
	}

	return s.Logout(c)
}

// Logout 退出登录
func (s *UserService) Logout(c *gin.Context) error {
	// 1、获取refresh_token
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		// 如果 Cookie 不存在或解析失败 (例如客户端没有发送，或者 Cookie 已过期)
		if errors.Is(err, http.ErrNoCookie) {
			global.CHAT_LOG.Warn("RefreshToken: Refresh token cookie not found.")
			return common.NewServiceError(common.REFRESH_TOKEN_INVALID)
		}
		// 其他获取 Cookie 的错误
		global.CHAT_LOG.Error("RefreshToken: 获取 refresh token cookie 失败", "err", err)
		return common.NewServiceError(common.ERROR)
	}
	refreshClaims := &utils.RefreshToken{}
	_, err = jwt.ParseWithClaims(refreshToken, refreshClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(global.CHAT_CONFIG.JWT.Secret), nil
	})

	// 2、从WebSocket连接池移除该用户所有连接
	manager := global.CHAT_WEBSOCKET_MANAGER.(*core.WebSocketManager)
	manager.UserLogoutByUserId(refreshClaims.UserID)

	// 3、撤销该用户token
	if err := utils.RevokeToken(refreshClaims.UserID, refreshClaims.Platform); err != nil {
		global.CHAT_LOG.Error("Logout-->撤销token失败", "err", err)
		return common.NewServiceError(common.ERROR)
	}

	// 4、清除cookie
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)

	return nil
}
