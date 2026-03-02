package service

import (
	"chat-server/constant"
	"chat-server/global"
	"chat-server/model"
	"chat-server/model/common"
	"chat-server/model/request/room"
	"chat-server/utils"
	"fmt"
	"strings"
)

type RoomService struct{}

// 创建房间
func (r *RoomService) CreateRoom(req room.CreateRoomRequest, userId string) (map[string]interface{}, error) {
	// 1、校验参数
	if !utils.VerifyString(req.RoomName) || !utils.VerifyString(req.Introduction) || !utils.VerifyString(req.Tag) || !utils.VerifyString(userId) {
		return nil, common.NewServiceError(common.INVALID_PARAMS)
	}
	if !constant.StatusMap[req.Status] {
		return nil, common.NewServiceError(common.INVALID_PARAMS)
	}
	if !utils.VerifyAvatar(req.Avatar) {
		return nil, common.NewServiceError(common.AVATAR_INVALID)
	}
	// 2、开启mysql事务
	tx := global.CHAT_MYSQL.Begin()
	if tx.Error != nil {
		global.CHAT_LOG.Error("CreateRoom-->开启Mysql事务失败", "err", tx.Error.Error())
		return nil, common.NewServiceError(common.ERROR)
	}
	defer func() {
		if r := recover(); r != nil {
			global.CHAT_LOG.Error("CreateRoom-->捕捉到panic", "err", r)
			tx.Rollback()
		} else if tx.Error != nil {
			global.CHAT_LOG.Error("CreateRoom-->捕捉到tx.Error", "err", r)
			tx.Rollback()
		} else {
			tx.Commit()
			global.CHAT_LOG.Info(fmt.Sprintf("CreateRoom-->%s-->mysql无报错", req.RoomName))
		}
	}()

	// 3、创建聊天室
	createRoom := model.Room{
		ID:           utils.GenerateUUid(),
		CreatorID:    userId,
		RoomName:     req.RoomName,
		Introduction: req.Introduction,
		Tag:          req.Tag,
		Status:       req.Status,
		Avatar:       req.Avatar,
		CreatedAt:    utils.GetUTCMillisTimestamp(),
		UpdatedAt:    utils.GetUTCMillisTimestamp(),
	}
	if err := tx.Create(&createRoom).Error; err != nil {
		tx.Error = err
		global.CHAT_LOG.Error("CreateRoom-->创建房间，数据库操作错误", "err", tx.Error.Error())
		return nil, common.NewServiceError(common.ERROR)
	}

	data := map[string]interface{}{
		"roomID": createRoom.ID,
	}
	return data, nil
}

// 搜索房间
func (r *RoomService) SearchRoom(req room.SearchRoomRequest) (map[string]interface{}, error) {
	// 1、校验参数
	if !utils.VerifyString(req.Content) {
		return nil, common.NewServiceError(common.INVALID_PARAMS)
	}

	// 2、开启mysql事务
	tx := global.CHAT_MYSQL.Begin()
	if tx.Error != nil {
		global.CHAT_LOG.Error("SearchRoom-->开启Mysql事务失败", "err", tx.Error.Error())
		return nil, common.NewServiceError(common.ERROR)
	}
	defer func() {
		if r := recover(); r != nil {
			global.CHAT_LOG.Error("SearchRoom-->捕捉到panic", "err", r)
			tx.Rollback()
		} else if tx.Error != nil {
			global.CHAT_LOG.Error("SearchRoom-->捕捉到tx.Error", "err", r)
			tx.Rollback()
		} else {
			tx.Commit()
			global.CHAT_LOG.Info(fmt.Sprintf("SearchRoom-->%s-->mysql无报错", req.Content))
		}
	}()

	// 3、搜索聊天室
	var rooms []model.Room
	switch req.Type {
	case 1:
		tx.Where("room_name LIKE ?", "%"+req.Content+"%").Find(&rooms)
	case 2:
		tx.Where("introduction LIKE ?", "%"+req.Content+"%").Find(&rooms)
	case 3: // 多个标签需要拆开查询
		tags := strings.Split(strings.Trim(req.Content, "#"), "#")
		for _, tag := range tags {
			tx = tx.Where("tag LIKE ?", "%#"+tag+"%")
		}
		tx.Distinct().Find(&rooms)
	}

	// 4、返回数据
	data := map[string]interface{}{
		"rooms": rooms,
	}
	return data, nil
}
