package cdn

// GetUpdateFileRoomItem 请求中的房间项
type GetUpdateFileRoomItem struct {
	RoomId    string `json:"room_id"`
	UpdatedAt int64  `json:"updated_at"`
}

// GetUpdateFileUserItem 请求中的用户项
type GetUpdateFileUserItem struct {
	UserId    string `json:"user_id"`
	UpdatedAt int64  `json:"updated_at"`
}

// GetUpdateFileRequest 获取更新资源请求
type GetUpdateFileRequest struct {
	Rooms []GetUpdateFileRoomItem `json:"rooms"`
	Users []GetUpdateFileUserItem `json:"users"`
}

// GetUpdateFileRoomResult 返回中的房间项
type GetUpdateFileRoomResult struct {
	RoomId string `json:"room_id"`
	Avatar string `json:"avatar"`
}

// GetUpdateFileUserResult 返回中的用户项
type GetUpdateFileUserResult struct {
	UserId string `json:"user_id"`
	Avatar string `json:"avatar"`
}
