package constant

const (
	// 用户
	MessageTypeText  = "text"  // 文本消息
	MessageTypeImage = "image" // 图片消息
	MessageTypeFile  = "file"  // 文件消息
	MessageTypeVoice = "voice" // 语音消息
	MessageTypeVideo = "video" // 视频消息
	MessageTypeReply = "reply" // 回复消息

	// 系统
	MessageTypeJoin       = "join"        // 加入房间
	MessageTypeLeave      = "leave"       // 离开房间
	MessageTypeRoomUpdate = "room_update" // 更新房间
	MessageTypeDelMember  = "del_member"  // 移除成员
	MessageTypeDelRoom    = "del_room"    // 删除房间
	MessageTypeSystem     = "system"      // 系统消息
	MessageTypeBan        = "ban"

	JoinMessageContent  = "已加入房间"
	LeaveMessageContent = "已离开房间"
)

var UserMessageType = map[string]bool{
	MessageTypeText:  true,
	MessageTypeImage: true,
	MessageTypeFile:  true,
	MessageTypeVoice: true,
	MessageTypeVideo: true,
	MessageTypeReply: true,
}

var SystemMessageType = map[string]bool{
	MessageTypeJoin:       true,
	MessageTypeLeave:      true,
	MessageTypeSystem:     true,
	MessageTypeRoomUpdate: true,
	MessageTypeDelMember:  true,
	MessageTypeDelRoom:    true,
}
