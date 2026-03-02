package service

import (
	"chat-server/constant"
	"chat-server/global"
	"chat-server/model"
	"chat-server/model/common"
	"chat-server/model/request/room"
	"chat-server/utils"
	"fmt"
)

type RoomService struct{}

func (r *RoomService) CreateRoom(req room.CreateRoomRequest, userId string) (string, error) {
	// 校验参数
	if !utils.VerifyString(req.RoomName) || !utils.VerifyString(req.Introduction) || !utils.VerifyString(req.Tag) || !utils.VerifyString(userId) {
		return "", common.NewServiceError(common.INVALID_PARAMS)
	}
	if !constant.StatusMap[req.Status] {
		return "", common.NewServiceError(common.INVALID_PARAMS)
	}
	// 开启mysql事务
	tx := global.CHAT_MYSQL.Begin()
	if tx.Error != nil {
		global.CHAT_LOG.Error("CreateRoom-->开启Mysql事务失败", "err", tx.Error.Error())
		return "", common.NewServiceError(common.ERROR)
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

	// 创建聊天室
	createRoom := model.Room{
		ID:           utils.GenerateUUid(),
		CreatorID:    userId,
		RoomName:     req.RoomName,
		Introduction: req.Introduction,
		Tag:          req.Tag,
		Status:       req.Status,
		CreatedAt:    utils.GetUTCMillisTimestamp(),
		UpdatedAt:    utils.GetUTCMillisTimestamp(),
	}
	if err := tx.Create(&createRoom).Error; err != nil {
		tx.Error = err
		global.CHAT_LOG.Error("CreateRoom-->创建房间，数据库操作错误", "err", tx.Error.Error())
		return "", common.NewServiceError(common.ERROR)
	}

	return createRoom.ID, nil
}
