package constant

const (
	MessageTypeText    = "text"    // 文本消息
	MessageTypeImage   = "image"   // 图片消息
	MessageTypeFile    = "file"    // 文件消息
	MessageTypeVoice   = "voice"   // 语音消息
	MessageTypeVideo   = "video"   // 视频消息
	MessageTypeReply   = "reply"   // 回复消息
	MessageTypeJoin    = "join"    // 加入房间
	MessageTypeLeave   = "leave"   // 离开房间
	MessageTypeSystem  = "system"  // 系统消息
	MessageTypeTyping  = "typing"  // 正在输入
	MessageTypeReceipt = "receipt" // 已读回执

	JoinMessageContent  = "用户已加入房间"
	LeaveMessageContent = "用户已离开房间"
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
	MessageTypeJoin:   true,
	MessageTypeLeave:  true,
	MessageTypeSystem: true,
}
