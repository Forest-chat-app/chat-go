package service

import (
	"chat-server/constant"
	"chat-server/core"
	"chat-server/global"
	"chat-server/model/common"
	"chat-server/model/mysql"
	"chat-server/model/request/room"
	"chat-server/utils"
	"fmt"
	"sort"
	"strings"
)

type RoomService struct{}

// CreateRoom 创建房间
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
	createRoom := mysql.Room{
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

	// 4、加入聊天室
	var queryRole mysql.Role
	tx.First(&queryRole, "role_id = ?", constant.RoleGroupOwner)
	if queryRole.ID == "" {
		return nil, common.NewServiceError(common.ROLE_NOT_FOUND)
	}
	var joinRoom = mysql.RoomMembers{
		ID:         utils.GenerateUUid(),
		UserID:     userId,
		RoomID:     createRoom.ID,
		JoinedAt:   utils.GetUTCMillisTimestamp(),
		LastReadId: "",
		UserRole:   queryRole.ID,
	}
	tx.Create(&joinRoom)

	// 5、返回数据
	data := map[string]interface{}{
		"roomID": createRoom.ID,
	}
	return data, nil
}

// SearchRoom 搜索房间
func (r *RoomService) SearchRoom(req room.SearchRoomRequest) (map[string]interface{}, error) {
	// 1、校验参数
	if !utils.VerifyString(req.Content) {
		return nil, common.NewServiceError(common.INVALID_PARAMS)
	}

	// 2、开启mysql事务
	tx := global.CHAT_MYSQL

	// 3、搜索聊天室
	var rooms []mysql.Room
	switch req.Type {
	case 1:
		tx.Where("room_name LIKE ? AND status = ?", "%"+req.Content+"%", constant.StatusPublic).Find(&rooms)
	case 2:
		tx.Where("introduction LIKE ? AND status = ?", "%"+req.Content+"%", constant.StatusPublic).Find(&rooms)
	case 3: // 多个标签需要拆开查询
		tags := strings.Split(strings.Trim(req.Content, "#"), "#")
		tx = tx.Where("status = ?", constant.StatusPublic)
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

// HotRoom 获取热门聊天室
func (r *RoomService) HotRoom() (map[string]interface{}, error) {
	// 1、获取所有公开房间
	tx := global.CHAT_MYSQL
	var rooms []mysql.Room
	tx.Where("status = ?", constant.StatusPublic).Find(&rooms)

	// 2、从WebSocket连接池获取每个房间的在线人数，并排序
	manager := global.CHAT_WEBSOCKET_MANAGER.(*core.WebSocketManager)
	onlineCounts := manager.GetRoomOnlineCounts()
	sort.Slice(rooms, func(i, j int) bool {
		return onlineCounts[rooms[i].ID] > onlineCounts[rooms[j].ID]
	})

	// 3、取前50个
	if len(rooms) > 50 {
		rooms = rooms[:50]
	}
	data := map[string]interface{}{
		"rooms": rooms,
	}
	return data, nil
}

// JoinRoom 加入房间
func (r *RoomService) JoinRoom(req room.JoinRoomRequest, userId string) (map[string]interface{}, error) {
	// 1、校验参数
	if !utils.VerifyString(req.RoomId) {
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
			global.CHAT_LOG.Info(fmt.Sprintf("SearchRoom-->%s-->mysql无报错", req.RoomId))
		}
	}()

	// 3、查询房间是否存在
	var queryRoom mysql.Room
	tx.First(&queryRoom, "id = ?", req.RoomId)
	if queryRoom.ID == "" {
		return nil, common.NewServiceError(common.ROOM_NOT_FOUND)
	}

	// 4、加入房间
	var queryRole mysql.Role
	tx.First(&queryRole, "role_id = ?", constant.RoleGroupMember)
	if queryRole.ID == "" {
		return nil, common.NewServiceError(common.ROLE_NOT_FOUND)
	}
	var joinRoom = mysql.RoomMembers{
		ID:         utils.GenerateUUid(),
		UserID:     userId,
		RoomID:     req.RoomId,
		JoinedAt:   utils.GetUTCMillisTimestamp(),
		LastReadId: "",
		UserRole:   queryRole.ID,
	}
	tx.Create(&joinRoom)

	// 5、发送加入消息
	joinMsg := &common.WebSocketMessage{
		Type:      constant.MessageTypeJoin,
		RoomId:    req.RoomId,
		SenderId:  userId,
		Content:   map[string]interface{}{constant.MessageTypeJoin: constant.JoinMessageContent, constant.MessageTypeLeave: nil, constant.MessageTypeSystem: nil},
		CreatedAt: utils.GetUTCMillisTimestamp(),
	}
	global.CHAT_WEBSOCKET_MANAGER.(*core.WebSocketManager).BroadcastToRoom(req.RoomId, joinMsg)
	return nil, nil
}

// QuitRoom 退出房间
func (r *RoomService) QuitRoom(req room.QuitRoomRequest, userId string) (map[string]interface{}, error) {
	// 1、校验参数
	if !utils.VerifyString(req.RoomId) {
		return nil, common.NewServiceError(common.INVALID_PARAMS)
	}

	// 2、开启mysql事务
	tx := global.CHAT_MYSQL.Begin()
	if tx.Error != nil {
		global.CHAT_LOG.Error("QuitRoom-->开启Mysql事务失败", "err", tx.Error.Error())
		return nil, common.NewServiceError(common.ERROR)
	}
	defer func() {
		if r := recover(); r != nil {
			global.CHAT_LOG.Error("QuitRoom-->捕捉到panic", "err", r)
			tx.Rollback()
		} else if tx.Error != nil {
			global.CHAT_LOG.Error("QuitRoom-->捕捉到tx.Error", "err", r)
			tx.Rollback()
		} else {
			tx.Commit()
			global.CHAT_LOG.Info(fmt.Sprintf("QuitRoom-->%s-->mysql无报错", req.RoomId))
		}
	}()

	// 3、查询房间是否存在
	var queryRoom mysql.Room
	tx.First(&queryRoom, "id = ?", req.RoomId)
	if queryRoom.ID == "" {
		return nil, common.NewServiceError(common.ROOM_NOT_FOUND)
	}

	// 4、退出房间
	var queryRoomMembers mysql.RoomMembers
	tx.First(&queryRoomMembers, "room_id = ? AND user_id = ?", req.RoomId, userId)
	if queryRoomMembers.ID == "" {
		return nil, common.NewServiceError(common.ROOM_NOT_FOUND)
	}

	// 5、退出房间
	tx.Delete(&queryRoomMembers)

	// 6、返回数据
	data := map[string]interface{}{
		"roomID": queryRoom.ID,
	}
	return data, nil
}

// PreviewRoom 预览房间
func (r *RoomService) PreviewRoom(req room.PreviewRoomRequest, userId string) {
	// 1、加入房间
	global.CHAT_WEBSOCKET_MANAGER.(*core.WebSocketManager).JoinRoom(req.RoomId, userId)
}

// LeavePreview 退出预览
func (r *RoomService) LeavePreview(req room.LeavePreviewRequest, userId string) {
	// 1、退出房间
	global.CHAT_WEBSOCKET_MANAGER.(*core.WebSocketManager).LeaveRoom(req.RoomId, userId)
}
