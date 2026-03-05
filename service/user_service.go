package service

import (
	"chat-server/constant"
	"chat-server/global"
	"chat-server/model/common"
	"chat-server/model/mysql"
	"chat-server/model/request/user"
	"chat-server/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	tokenPair, err := utils.GenerateTokenPair(userID, req.UserAccount)
	if err != nil {
		tx.Error = err
		global.CHAT_LOG.Error("RegisterUser-->生成token失败", "err", err)
		return nil, common.NewServiceError(common.GENERATE_TOKEN_ERROR)
	}
	// 5.1、在redis保存RefreshToken状态
	tokenId := utils.GenerateUUid()
	err = utils.StoreRefreshToken(userID, tokenId, req.Platform)
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
	redis := global.CHAT_REDIS
	ctx := context.Background()

	// 对mysql事务进行操作
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

	// 验证userAccount
	var queryUser mysql.User
	err := tx.Where("user_account = ?", req.UserAccount).First(&queryUser).Error
	if err != nil {
		global.CHAT_LOG.Error("LoginAccount-->检查用户账号，数据库操作错误", "err", err)
		return nil, common.NewServiceError(common.USER_ACCOUNT_NOT_FOUND)
	}

	// 验证密码
	match, err := utils.CompareHashAndPassword(queryUser.Password, req.Password)
	if err != nil {
		return nil, common.NewServiceError(common.ERROR)
	}
	if !match {
		return nil, common.NewServiceError(common.PASSWORD_INVALID)
	}

	// 检查redis是否已存在该登录平台的token，只允许单平台登录
	tokenKey := fmt.Sprintf("user_tokens:%s", queryUser.ID)
	tokenIds, err := redis.SMembers(ctx, tokenKey).Result()
	if err != nil {
		global.CHAT_LOG.Error("LoginAccount-->检查该用户所有tokenId失败", "err", err)
		return nil, common.NewServiceError(common.ERROR)
	}
	if len(tokenIds) > 0 {
		for _, tokenId := range tokenIds {
			// 通过tokenId获取refreshToken
			refreshToken := fmt.Sprintf("refresh_token:%s:%s", queryUser.ID, tokenId)
			refreshTokenData, err := redis.Get(ctx, refreshToken).Result()
			if err != nil {
				global.CHAT_LOG.Error("LoginAccount-->获取refreshToken失败", "err", err)
				return nil, common.NewServiceError(common.ERROR)
			}
			// 解析值，获取登录平台信息
			var tokenData map[string]interface{}
			if err = json.Unmarshal([]byte(refreshTokenData), &tokenData); err != nil {
				global.CHAT_LOG.Error("LoginAccount-->解析refreshTokenData失败", "err", err)
				return nil, common.NewServiceError(common.ERROR)
			}
			// 平台相同则撤销旧令牌
			getPlatform, ok := tokenData["platform"].(string)
			if !ok {
				global.CHAT_LOG.Error("LoginAccount-->获取platform失败", "err", err)
				return nil, common.NewServiceError(common.ERROR)
			}
			if getPlatform == req.Platform {
				err := utils.RevokeToken(queryUser.ID, tokenId)
				if err != nil {
					global.CHAT_LOG.Error("LoginAccount-->撤销旧令牌RevokeToken失败", "err", err)
					return nil, err
				}
			}

		}
	}
	// 通过验证，下发token
	tokenPair, err := utils.GenerateTokenPair(queryUser.ID, queryUser.UserAccount)
	if err != nil {
		global.CHAT_LOG.Error("LoginAccount-->生成token失败", "err", err)
		return nil, common.NewServiceError(common.GENERATE_TOKEN_ERROR)
	}

	// 在redis保存RefreshToken状态
	tokenId := uuid.New().String()
	err = utils.StoreRefreshToken(queryUser.ID, tokenId, req.Platform)
	if err != nil {
		global.CHAT_LOG.Error("LoginAccount-->保存RefreshToken状态失败", "err", err)
		return nil, common.NewServiceError(common.ERROR)
	}

	// 处理返回数据
	// 设置refresh_token为http only
	cookieName := "refresh_token"
	maxAge := int(time.Duration(global.CHAT_CONFIG.JWT.RefreshTime) * 24 * time.Hour / time.Second)
	// 设置 Cookie
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
