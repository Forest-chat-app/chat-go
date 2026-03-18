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
		Avatar:       constant.RoomAvatar,
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

	// 5、加入到ws连接中
	ws := global.CHAT_WEBSOCKET_MANAGER.(*core.WebSocketManager)
	ws.JoinRoom(createRoom.ID, userId)

	// 6、返回数据
	data := map[string]interface{}{
		"roomID": createRoom.ID,
	}
	return data, nil
}

// GetRoomInfo 获取房间信息
func (r *RoomService) GetRoomInfo(req room.GetRoomInfoRequest) (map[string]interface{}, error) {
	if !utils.VerifyString(req.RoomId) {
		return nil, common.NewServiceError(common.INVALID_PARAMS)
	}

	var queryRoom mysql.Room
	global.CHAT_MYSQL.First(&queryRoom, "id = ?", req.RoomId)
	if queryRoom.ID == "" {
		return nil, common.NewServiceError(common.ROOM_NOT_FOUND)
	}
	queryRoom.Avatar = utils.GenerateCdnUrl(queryRoom.Avatar)

	return map[string]interface{}{
		"room": queryRoom,
	}, nil
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
	// 3.1 获取头像url
	for i := range rooms {
		rooms[i].Avatar = utils.GenerateCdnUrl(rooms[i].Avatar)
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

	// 2、从WebSocket连接池获取每个房间的在线人数
	manager := global.CHAT_WEBSOCKET_MANAGER.(*core.WebSocketManager)
	onlineCounts := manager.GetRoomOnlineCounts()

	// 3、构建带在线人数的房间列表
	roomsWithCount := make([]room.RoomWithOnlineCount, len(rooms))
	for i, r := range rooms {
		roomsWithCount[i] = room.RoomWithOnlineCount{
			Room:        r,
			OnlineCount: onlineCounts[r.ID],
		}
	}

	// 4、按在线人数排序
	sort.Slice(roomsWithCount, func(i, j int) bool {
		return roomsWithCount[i].OnlineCount > roomsWithCount[j].OnlineCount
	})

	// 5、取前50个
	if len(roomsWithCount) > 50 {
		roomsWithCount = roomsWithCount[:50]
	}
	for i := range roomsWithCount {
		roomsWithCount[i].Avatar = utils.GenerateCdnUrl(roomsWithCount[i].Avatar)
	}
	data := map[string]interface{}{
		"rooms": roomsWithCount,
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
	queryUser, _ := utils.GetUserByID(userId)
	joinMsg := &common.WebSocketMessage{
		Type:            constant.MessageTypeJoin,
		RoomId:          req.RoomId,
		SenderId:        userId,
		SenderUpdatedAt: 0, // 非聊天消息，不进行用户对比
		Content:         map[string]interface{}{"text": queryUser.Nickname + constant.JoinMessageContent},
		CreatedAt:       utils.GetUTCMillisTimestamp(),
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
	tx.Delete(&queryRoomMembers)

	// 5、断开连接
	global.CHAT_WEBSOCKET_MANAGER.(*core.WebSocketManager).LeaveRoom(queryRoom.ID, userId)

	// 6、发送退出消息
	leaveMsg := &common.WebSocketMessage{
		Type:            constant.MessageTypeLeave,
		RoomId:          req.RoomId,
		SenderId:        userId,
		SenderUpdatedAt: 0, // 非聊天消息，不进行用户对比
		Content:         map[string]interface{}{"userId": userId},
		CreatedAt:       utils.GetUTCMillisTimestamp(),
	}
	global.CHAT_WEBSOCKET_MANAGER.(*core.WebSocketManager).BroadcastToRoom(req.RoomId, leaveMsg)

	// 7、返回数据
	data := map[string]interface{}{
		"roomID": queryRoom.ID,
	}
	return data, nil
}

// DeleteRoomMembers 移除房间成员
func (r *RoomService) DeleteRoomMembers(req room.DeleteRoomMembersRequest, userId string) error {
	// 1、校验参数
	if !utils.VerifyString(req.RoomId) || len(req.UserIds) == 0 {
		return common.NewServiceError(common.INVALID_PARAMS)
	}

	tx := global.CHAT_MYSQL

	// 2、多表联查获取当前用户在该房间的 role_id，判断是否有权限
	var roleId int16
	err := tx.Table("room_members rm").
		Select("r.role_id").
		Joins("JOIN role r ON r.id = rm.user_role").
		Where("rm.user_id = ? AND rm.room_id = ?", userId, req.RoomId).
		Scan(&roleId).Error
	if err != nil || roleId == 0 {
		return common.NewServiceError(common.ROOM_PERMISSION_DENY)
	}
	if roleId != constant.RoleGroupOwner && roleId != constant.RoleGroupAdmin {
		return common.NewServiceError(common.ROOM_PERMISSION_DENY)
	}

	// 3、批量删除目标用户在该房间的成员记录
	if err := tx.Where("room_id = ? AND user_id IN ?", req.RoomId, req.UserIds).
		Delete(&mysql.RoomMembers{}).Error; err != nil {
		global.CHAT_LOG.Error("DeleteRoomMembers-->删除成员失败", "err", err.Error())
		return common.NewServiceError(common.ERROR)
	}

	// 4、发送移除消息
	delMemberMsg := &common.WebSocketMessage{
		Type:            constant.MessageTypeDelMember,
		RoomId:          req.RoomId,
		SenderId:        userId,
		SenderUpdatedAt: 0, // 非聊天消息，不进行用户对比
		Content:         map[string]interface{}{"userIds": req.UserIds},
		CreatedAt:       utils.GetUTCMillisTimestamp(),
	}
	global.CHAT_WEBSOCKET_MANAGER.(*core.WebSocketManager).BroadcastToRoom(req.RoomId, delMemberMsg)

	return nil
}

// GetRoomMembers 获取房间成员列表
func (r *RoomService) GetRoomMembers(req room.GetRoomMembersRequest) (map[string]interface{}, error) {
	// 1、校验参数
	if !utils.VerifyString(req.RoomId) {
		return nil, common.NewServiceError(common.INVALID_PARAMS)
	}

	// 2、查询房间是否存在
	tx := global.CHAT_MYSQL
	var queryRoom mysql.Room
	tx.First(&queryRoom, "id = ?", req.RoomId)
	if queryRoom.ID == "" {
		return nil, common.NewServiceError(common.ROOM_NOT_FOUND)
	}

	// 3、查询房间成员
	var roomMembers []mysql.RoomMembers
	tx.Where("room_id = ?", req.RoomId).Find(&roomMembers)

	// 4、获取每个成员的用户信息
	members := make([]room.RoomMember, 0, len(roomMembers))
	for _, member := range roomMembers {
		var user mysql.User
		tx.Select("id, nickname, avatar, email").First(&user, "id = ?", member.UserID)
		if user.ID == "" {
			continue
		}
		members = append(members, room.RoomMember{
			UserID:   user.ID,
			Nickname: user.Nickname,
			Avatar:   utils.GenerateCdnUrl(user.Avatar),
			Email:    user.Email,
			JoinedAt: member.JoinedAt,
		})
	}

	// 5、返回数据
	data := map[string]interface{}{
		"members": members,
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

// DeleteRoom 解散房间
func (r *RoomService) DeleteRoom(req room.DeleteRoomRequest, userId string) error {
	// 1、校验参数
	if !utils.VerifyString(req.RoomId) {
		return common.NewServiceError(common.INVALID_PARAMS)
	}

	// 2、查询房间，验证是否是创建者
	tx := global.CHAT_MYSQL
	var queryRoom mysql.Room
	tx.Select("id, creator_id").First(&queryRoom, "id = ?", req.RoomId)
	if queryRoom.ID == "" {
		return common.NewServiceError(common.ROOM_NOT_FOUND)
	}
	if queryRoom.CreatorID != userId {
		return common.NewServiceError(common.ROOM_PERMISSION_DENY)
	}

	// 3、开启事务
	txDb := global.CHAT_MYSQL.Begin()
	if txDb.Error != nil {
		global.CHAT_LOG.Error("DeleteRoom-->开启Mysql事务失败", "err", txDb.Error.Error())
		return common.NewServiceError(common.ERROR)
	}
	defer func() {
		if r := recover(); r != nil {
			global.CHAT_LOG.Error("DeleteRoom-->捕捉到panic", "err", r)
			txDb.Rollback()
		} else if txDb.Error != nil {
			global.CHAT_LOG.Error("DeleteRoom-->捕捉到tx.Error", "err", txDb.Error)
			txDb.Rollback()
		} else {
			txDb.Commit()
		}
	}()

	// 4、将 room 状态改为 StatusDelete
	if err := txDb.Model(&mysql.Room{}).Where("id = ?", req.RoomId).Updates(map[string]interface{}{
		"status":     constant.StatusDelete,
		"updated_at": utils.GetUTCMillisTimestamp(),
	}).Error; err != nil {
		txDb.Error = err
		global.CHAT_LOG.Error("DeleteRoom-->更新房间状态失败", "err", err.Error())
		return common.NewServiceError(common.ERROR)
	}

	// 5、删除 room_members 表中所有该房间的成员记录
	if err := txDb.Where("room_id = ?", req.RoomId).Delete(&mysql.RoomMembers{}).Error; err != nil {
		txDb.Error = err
		global.CHAT_LOG.Error("DeleteRoom-->删除房间成员失败", "err", err.Error())
		return common.NewServiceError(common.ERROR)
	}

	// 6、广播解散通知
	delRoomMsg := &common.WebSocketMessage{
		Type:            constant.MessageTypeDelRoom,
		RoomId:          req.RoomId,
		SenderId:        userId,
		SenderUpdatedAt: 0, // 非聊天消息，不进行用户对比
		Content:         map[string]interface{}{},
		CreatedAt:       utils.GetUTCMillisTimestamp(),
	}
	global.CHAT_WEBSOCKET_MANAGER.(*core.WebSocketManager).BroadcastToRoom(req.RoomId, delRoomMsg)

	// 7、从 WebSocket 管理器中删除房间
	global.CHAT_WEBSOCKET_MANAGER.(*core.WebSocketManager).DelRoom(req.RoomId)

	return nil
}

// UpdateRoomInfo 更新房间信息
func (r *RoomService) UpdateRoomInfo(req room.UpdateRoomInfoRequest, userId string) error {
	// 1、校验参数
	if !utils.VerifyString(req.RoomId) {
		return common.NewServiceError(common.INVALID_PARAMS)
	}

	tx := global.CHAT_MYSQL

	// 2、查询房间，验证是否是创建者
	var queryRoom mysql.Room
	tx.Select("creator_id").First(&queryRoom, "id = ?", req.RoomId)
	if queryRoom.CreatorID == "" {
		return common.NewServiceError(common.ROOM_NOT_FOUND)
	}
	if queryRoom.CreatorID != userId {
		return common.NewServiceError(common.ROOM_PERMISSION_DENY)
	}

	// 3、构造更新字段（只更新非空字段）
	updates := map[string]interface{}{
		"updated_at": utils.GetUTCMillisTimestamp(),
	}
	if utils.VerifyString(req.RoomName) {
		updates["room_name"] = req.RoomName
	}
	if utils.VerifyString(req.Introduction) {
		updates["introduction"] = req.Introduction
	}
	if utils.VerifyString(req.Tag) {
		updates["tag"] = req.Tag
	}
	if req.Status != 0 {
		updates["status"] = req.Status
	}

	if err := tx.Model(&mysql.Room{}).Where("id = ?", req.RoomId).Updates(updates).Error; err != nil {
		global.CHAT_LOG.Error("UpdateRoomInfo-->更新房间信息失败", "err", err.Error())
		return common.NewServiceError(common.ERROR)
	}

	// 4、广播房间更新通知
	updateRoomMsg := &common.WebSocketMessage{
		Type:            constant.MessageTypeRoomUpdate,
		RoomId:          req.RoomId,
		SenderId:        userId,
		SenderUpdatedAt: 0, // 非聊天消息，不进行用户对比
		Content:         map[string]interface{}{},
		CreatedAt:       utils.GetUTCMillisTimestamp(),
	}
	global.CHAT_WEBSOCKET_MANAGER.(*core.WebSocketManager).BroadcastToRoom(req.RoomId, updateRoomMsg)

	return nil
}

// UpdateRoomAvatar 更新房间头像
func (r *RoomService) UpdateRoomAvatar(req room.UpdateRoomAvatarRequest, userId string) error {
	// 1、开启Mysql事务
	tx := global.CHAT_MYSQL.Begin()
	if tx.Error != nil {
		global.CHAT_LOG.Error("UpdateRoomAvatar-->开启Mysql事务失败", "err", tx.Error.Error())
		return common.NewServiceError(common.ERROR)
	}
	defer func() {
		if r := recover(); r != nil {
			global.CHAT_LOG.Error("UpdateRoomAvatar-->捕捉到panic", "err", r)
			tx.Rollback()
		} else if tx.Error != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// 2、更新房间头像
	if err := tx.Model(&mysql.Room{}).Where("id = ?", req.RoomId).Updates(map[string]interface{}{
		"avatar":     req.Avatar,
		"updated_at": utils.GetUTCMillisTimestamp(),
	}).Error; err != nil {
		tx.Error = err
		global.CHAT_LOG.Error("UpdateRoomAvatar-->更新用户头像失败", "err", err)
		return common.NewServiceError(common.ERROR)
	}

	// 3、广播房间更新通知
	updateRoomMsg := &common.WebSocketMessage{
		Type:            constant.MessageTypeRoomUpdate,
		RoomId:          req.RoomId,
		SenderId:        userId,
		SenderUpdatedAt: 0, // 非聊天消息，不进行用户对比
		Content:         map[string]interface{}{"avatar": utils.GenerateCdnUrl(req.Avatar)},
		CreatedAt:       utils.GetUTCMillisTimestamp(),
	}
	global.CHAT_WEBSOCKET_MANAGER.(*core.WebSocketManager).BroadcastToRoom(req.RoomId, updateRoomMsg)

	return nil
}
