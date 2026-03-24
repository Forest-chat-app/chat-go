package admin

// ListRoomsRequest 房间列表请求
type ListRoomsRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Keyword  string `form:"keyword"`
	Status   int16  `form:"status"` // 0=全部 1=公开 2=私密 3=已删除
}

// AdminCreateRoomRequest 管理端创建房间
type AdminCreateRoomRequest struct {
	RoomName     string `json:"room_name" binding:"required"`
	Introduction string `json:"introduction" binding:"required"`
	Tag          string `json:"tag" binding:"required"`
	Status       int16  `json:"status" binding:"required"`
	CreatorID    string `json:"creator_id" binding:"required"`
}

// AdminUpdateRoomInfoRequest 管理端修改房间信息（不含头像）
type AdminUpdateRoomInfoRequest struct {
	RoomId       string `json:"room_id" binding:"required"`
	RoomName     string `json:"room_name"`
	Introduction string `json:"introduction"`
	Tag          string `json:"tag"`
	Status       int16  `json:"status"`
}

// AdminUpdateRoomAvatarRequest 管理端修改房间头像
type AdminUpdateRoomAvatarRequest struct {
	RoomId string `json:"room_id" binding:"required"`
	Avatar string `json:"avatar" binding:"required"`
}

// AdminDeleteRoomRequest 管理端解散房间
type AdminDeleteRoomRequest struct {
	RoomId string `json:"room_id" binding:"required"`
}

// GetRoomDetailRequest 获取房间详情
type GetRoomDetailRequest struct {
	RoomId string `form:"room_id" binding:"required"`
}

// AdminGetRoomMembersRequest 获取房间成员列表
type AdminGetRoomMembersRequest struct {
	RoomId string `form:"room_id" binding:"required"`
}

// AdminRemoveRoomMemberRequest 移除房间成员
type AdminRemoveRoomMemberRequest struct {
	RoomId  string   `json:"room_id" binding:"required"`
	UserIds []string `json:"user_ids" binding:"required,min=1"`
}
