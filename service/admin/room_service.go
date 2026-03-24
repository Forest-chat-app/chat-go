package admin

import (
	"chat-server/constant"
	"chat-server/core"
	"chat-server/global"
	"chat-server/model/common"
	"chat-server/model/mysql"
	reqAdmin "chat-server/model/request/admin"
	"chat-server/model/request/room"
	"chat-server/utils"
)

type AdminRoomService struct{}

// ListRooms 房间列表（分页+关键词+状态筛选）
func (s *AdminRoomService) ListRooms(req reqAdmin.ListRoomsRequest) (map[string]interface{}, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	tx := global.CHAT_MYSQL.Model(&mysql.Room{})
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		tx = tx.Where("room_name LIKE ? OR introduction LIKE ?", like, like)
	}
	if req.Status != 0 {
		tx = tx.Where("status = ?", req.Status)
	}

	var total int64
	tx.Count(&total)

	var rooms []mysql.Room
	tx.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&rooms)

	for i := range rooms {
		rooms[i].Avatar = utils.GenerateCdnUrl(rooms[i].Avatar)
	}

	return map[string]interface{}{
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
		"rooms":     rooms,
	}, nil
}

// GetRoomDetail 获取房间详情
func (s *AdminRoomService) GetRoomDetail(roomId string) (map[string]interface{}, error) {
	var queryRoom mysql.Room
	if err := global.CHAT_MYSQL.First(&queryRoom, "id = ?", roomId).Error; err != nil {
		return nil, common.NewServiceError(common.ROOM_NOT_FOUND)
	}
	queryRoom.Avatar = utils.GenerateCdnUrl(queryRoom.Avatar)

	var memberCount int64
	global.CHAT_MYSQL.Model(&mysql.RoomMembers{}).Where("room_id = ?", roomId).Count(&memberCount)

	return map[string]interface{}{
		"room":         queryRoom,
		"member_count": memberCount,
	}, nil
}

// CreateRoom 管理端创建房间
func (s *AdminRoomService) CreateRoom(req reqAdmin.AdminCreateRoomRequest) (map[string]interface{}, error) {
	if !constant.StatusMap[req.Status] {
		return nil, common.NewServiceError(common.INVALID_PARAMS)
	}
	// 验证创建者存在
	var creator mysql.User
	if err := global.CHAT_MYSQL.First(&creator, "id = ?", req.CreatorID).Error; err != nil {
		return nil, common.NewServiceError(common.USER_ID_NOT_FOUND)
	}

	tx := global.CHAT_MYSQL.Begin()
	if tx.Error != nil {
		return nil, common.NewServiceError(common.ERROR)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		} else if tx.Error != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	newRoom := mysql.Room{
		ID:           utils.GenerateUUid(),
		CreatorID:    req.CreatorID,
		RoomName:     req.RoomName,
		Introduction: req.Introduction,
		Tag:          req.Tag,
		Status:       req.Status,
		Avatar:       constant.RoomAvatar,
		CreatedAt:    utils.GetUTCMillisTimestamp(),
		UpdatedAt:    utils.GetUTCMillisTimestamp(),
	}
	if err := tx.Create(&newRoom).Error; err != nil {
		tx.Error = err
		return nil, common.NewServiceError(common.ERROR)
	}

	// 将创建者加入房间（群主角色）
	var ownerRole mysql.Role
	tx.First(&ownerRole, "role_id = ?", constant.RoleGroupOwner)
	if ownerRole.ID != "" {
		tx.Create(&mysql.RoomMembers{
			ID:         utils.GenerateUUid(),
			UserID:     req.CreatorID,
			RoomID:     newRoom.ID,
			JoinedAt:   utils.GetUTCMillisTimestamp(),
			LastReadId: "",
			UserRole:   ownerRole.ID,
		})
	}

	// 加入ws
	global.CHAT_WEBSOCKET_MANAGER.(*core.WebSocketManager).JoinRoom(newRoom.ID, req.CreatorID)

	return map[string]interface{}{"room_id": newRoom.ID}, nil
}

// UpdateRoomInfo 修改房间基本信息（不含头像）
func (s *AdminRoomService) UpdateRoomInfo(req reqAdmin.AdminUpdateRoomInfoRequest) error {
	var queryRoom mysql.Room
	if err := global.CHAT_MYSQL.Select("id").First(&queryRoom, "id = ?", req.RoomId).Error; err != nil {
		return common.NewServiceError(common.ROOM_NOT_FOUND)
	}

	updates := map[string]interface{}{
		"updated_at": utils.GetUTCMillisTimestamp(),
	}
	if req.RoomName != "" {
		updates["room_name"] = req.RoomName
	}
	if req.Introduction != "" {
		updates["introduction"] = req.Introduction
	}
	if req.Tag != "" {
		updates["tag"] = req.Tag
	}
	if req.Status != 0 {
		updates["status"] = req.Status
	}

	if err := global.CHAT_MYSQL.Model(&mysql.Room{}).Where("id = ?", req.RoomId).Updates(updates).Error; err != nil {
		global.CHAT_LOG.Error("AdminUpdateRoomInfo-->更新失败", "err", err)
		return common.NewServiceError(common.ERROR)
	}

	// 广播更新通知
	updateMsg := &common.WebSocketMessage{
		Type:            constant.MessageTypeRoomUpdate,
		RoomId:          req.RoomId,
		SenderId:        "admin",
		SenderUpdatedAt: 0,
		Content:         map[string]interface{}{},
		CreatedAt:       utils.GetUTCMillisTimestamp(),
	}
	global.CHAT_WEBSOCKET_MANAGER.(*core.WebSocketManager).BroadcastToRoom(req.RoomId, updateMsg)
	return nil
}

// UpdateRoomAvatar 修改房间头像
func (s *AdminRoomService) UpdateRoomAvatar(req reqAdmin.AdminUpdateRoomAvatarRequest) error {
	var queryRoom mysql.Room
	if err := global.CHAT_MYSQL.Select("id").First(&queryRoom, "id = ?", req.RoomId).Error; err != nil {
		return common.NewServiceError(common.ROOM_NOT_FOUND)
	}

	if err := global.CHAT_MYSQL.Model(&mysql.Room{}).Where("id = ?", req.RoomId).Updates(map[string]interface{}{
		"avatar":     req.Avatar,
		"updated_at": utils.GetUTCMillisTimestamp(),
	}).Error; err != nil {
		return common.NewServiceError(common.ERROR)
	}

	updateMsg := &common.WebSocketMessage{
		Type:            constant.MessageTypeRoomUpdate,
		RoomId:          req.RoomId,
		SenderId:        "admin",
		SenderUpdatedAt: 0,
		Content:         map[string]interface{}{"avatar": utils.GenerateCdnUrl(req.Avatar)},
		CreatedAt:       utils.GetUTCMillisTimestamp(),
	}
	global.CHAT_WEBSOCKET_MANAGER.(*core.WebSocketManager).BroadcastToRoom(req.RoomId, updateMsg)
	return nil
}

// DeleteRoom 解散房间
func (s *AdminRoomService) DeleteRoom(req reqAdmin.AdminDeleteRoomRequest) error {
	var queryRoom mysql.Room
	if err := global.CHAT_MYSQL.Select("id").First(&queryRoom, "id = ?", req.RoomId).Error; err != nil {
		return common.NewServiceError(common.ROOM_NOT_FOUND)
	}

	tx := global.CHAT_MYSQL.Begin()
	if tx.Error != nil {
		return common.NewServiceError(common.ERROR)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		} else if tx.Error != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	if err := tx.Model(&mysql.Room{}).Where("id = ?", req.RoomId).Updates(map[string]interface{}{
		"status":     constant.StatusDelete,
		"updated_at": utils.GetUTCMillisTimestamp(),
	}).Error; err != nil {
		tx.Error = err
		return common.NewServiceError(common.ERROR)
	}

	if err := tx.Where("room_id = ?", req.RoomId).Delete(&mysql.RoomMembers{}).Error; err != nil {
		tx.Error = err
		return common.NewServiceError(common.ERROR)
	}

	// 广播解散通知
	delMsg := &common.WebSocketMessage{
		Type:            constant.MessageTypeDelRoom,
		RoomId:          req.RoomId,
		SenderId:        "admin",
		SenderUpdatedAt: 0,
		Content:         map[string]interface{}{},
		CreatedAt:       utils.GetUTCMillisTimestamp(),
	}
	global.CHAT_WEBSOCKET_MANAGER.(*core.WebSocketManager).BroadcastToRoom(req.RoomId, delMsg)
	global.CHAT_WEBSOCKET_MANAGER.(*core.WebSocketManager).DelRoom(req.RoomId)
	return nil
}

// GetRoomMembers 获取房间成员列表
func (s *AdminRoomService) GetRoomMembers(req reqAdmin.AdminGetRoomMembersRequest) (map[string]interface{}, error) {
	var queryRoom mysql.Room
	if err := global.CHAT_MYSQL.Select("id").First(&queryRoom, "id = ?", req.RoomId).Error; err != nil {
		return nil, common.NewServiceError(common.ROOM_NOT_FOUND)
	}

	var roomMembers []mysql.RoomMembers
	global.CHAT_MYSQL.Where("room_id = ?", req.RoomId).Find(&roomMembers)

	members := make([]room.RoomMember, 0, len(roomMembers))
	for _, member := range roomMembers {
		var user mysql.User
		global.CHAT_MYSQL.Select("id, nickname, avatar, email, created_at, updated_at").First(&user, "id = ?", member.UserID)
		if user.ID == "" {
			continue
		}
		members = append(members, room.RoomMember{
			UserID:    user.ID,
			Nickname:  user.Nickname,
			Avatar:    utils.GenerateCdnUrl(user.Avatar),
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			JoinedAt:  member.JoinedAt,
		})
	}

	return map[string]interface{}{"members": members}, nil
}

// RemoveRoomMember 移除房间成员
func (s *AdminRoomService) RemoveRoomMember(req reqAdmin.AdminRemoveRoomMemberRequest) error {
	if err := global.CHAT_MYSQL.Where("room_id = ? AND user_id IN ?", req.RoomId, req.UserIds).
		Delete(&mysql.RoomMembers{}).Error; err != nil {
		global.CHAT_LOG.Error("AdminRemoveRoomMember-->删除失败", "err", err)
		return common.NewServiceError(common.ERROR)
	}

	delMsg := &common.WebSocketMessage{
		Type:            constant.MessageTypeDelMember,
		RoomId:          req.RoomId,
		SenderId:        "admin",
		SenderUpdatedAt: 0,
		Content:         map[string]interface{}{"userIds": req.UserIds},
		CreatedAt:       utils.GetUTCMillisTimestamp(),
	}
	global.CHAT_WEBSOCKET_MANAGER.(*core.WebSocketManager).BroadcastToRoom(req.RoomId, delMsg)
	return nil
}
