package service

import (
	"chat-server/global"
	"chat-server/middleware"
	"chat-server/model"
	"chat-server/model/common"
	"chat-server/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
)

type UserService struct{}

// RegisterUser 注册用户
func (s *UserService) RegisterUser(userAccount, password, email, platform string) (*middleware.TokenPair, error) {
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
			global.CHAT_LOG.Info(fmt.Sprintf("RegisterUser-->%s 成功", userAccount))
		}
	}()
	// 检查用户名是否已存在
	var count int64
	err := tx.Model(&model.User{}).Where("user_account = ?", userAccount).Count(&count).Error
	if err != nil {
		tx.Error = err
		global.CHAT_LOG.Error("RegisterUser-->检查用户账号，数据库操作错误", "err", err)
		return nil, common.NewServiceError(common.ERROR)
	}
	if count > 0 {
		return nil, common.NewServiceError(common.USER_ACCOUNT_EXISTS)
	}

	// 判断密码是否合法
	if len(password) <= 0 {
		return nil, common.NewServiceError(common.PASSWORD_INVALID)
	}
	hashedPassword, err := utils.GenerateFromPassword(password)
	if err != nil || hashedPassword == "" {
		global.CHAT_LOG.Error("RegisterUser-->加密密码出错", "err", err)
		return nil, common.NewServiceError(common.ERROR)
	}

	// 判断邮箱是否合法
	if !utils.VerifyEmail(email) {
		return nil, common.NewServiceError(common.EMAIL_INVALID)
	}

	// 创建新用户
	userID := uuid.New().String()
	user := model.User{
		ID:          userID,
		UserAccount: userAccount,
		Password:    hashedPassword,
		Nickname:    userAccount,
		Email:       email,
		Avatar:      "",
		CreatedAt:   utils.GetUTCMillisTimestamp(),
		UpdatedAt:   utils.GetUTCMillisTimestamp(),
	}
	err = tx.Create(&user).Error
	if err != nil {
		tx.Error = err
		global.CHAT_LOG.Error("RegisterUser-->创建用户，数据库操作错误", "err", err)
		return nil, common.NewServiceError(common.ERROR)
	}

	// 创建用户成功, 生成token
	tokenPair, err := utils.GenerateTokenPair(userID, userAccount)
	if err != nil {
		tx.Error = err
		global.CHAT_LOG.Error("RegisterUser-->生成token失败", "err", err)
		return nil, common.NewServiceError(common.GENERATE_TOKEN_ERROR)
	}
	// 在redis保存RefreshToken状态
	tokenId := uuid.New().String()
	err = utils.StoreRefreshToken(userID, tokenId, platform)
	if err != nil {
		tx.Error = err
		global.CHAT_LOG.Error("RegisterUser-->保存RefreshToken状态失败", "err", err)
		return nil, common.NewServiceError(common.ERROR)
	}

	return tokenPair, nil
}

func (s *UserService) LoginAccount(userAccount string, password string, platform string) (*middleware.TokenPair, error) {
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
			global.CHAT_LOG.Info(fmt.Sprintf("LoginAccount-->%s 成功", userAccount))
		}
	}()

	// 验证userAccount
	var queryUser model.User
	err := tx.Where("user_account = ?", userAccount).First(&queryUser).Error
	if err != nil {
		global.CHAT_LOG.Error("LoginAccount-->检查用户账号，数据库操作错误", "err", err)
		return nil, common.NewServiceError(common.USER_ACCOUNT_NOT_FOUND)
	}

	// 验证密码
	match, err := utils.CompareHashAndPassword(queryUser.Password, password)
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
			if getPlatform == platform {
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
	err = utils.StoreRefreshToken(queryUser.ID, tokenId, platform)
	if err != nil {
		global.CHAT_LOG.Error("LoginAccount-->保存RefreshToken状态失败", "err", err)
		return nil, common.NewServiceError(common.ERROR)
	}

	return tokenPair, nil
}
