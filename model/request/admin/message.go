package admin

// AdminListMessagesRequest 管理端列出房间消息
type AdminListMessagesRequest struct {
	RoomId   string `form:"room_id" binding:"required"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// AdminDeleteMessageRequest 管理端删除消息
type AdminDeleteMessageRequest struct {
	MessageId string `json:"message_id" binding:"required"` // MongoDB ObjectID
}

// AdminSearchMessagesRequest 管理端全局搜索消息
type AdminSearchMessagesRequest struct {
	Keyword  string `form:"keyword" binding:"required"`
	RoomId   string `form:"room_id"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// AdminListAdminsRequest 列出管理员
type AdminListAdminsRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// AdminCreateAdminRequest 超级管理员创建普通管理员
type AdminCreateAdminRequest struct {
	UserId string `json:"user_id" binding:"required"` // 将该普通用户提升为管理员
}

// AdminDeleteAdminRequest 超级管理员删除普通管理员
type AdminDeleteAdminRequest struct {
	UserId string `json:"user_id" binding:"required"`
}
