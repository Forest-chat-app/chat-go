package room

import "chat-server/model/mysql"

// RoomWithOnlineCount 房间信息（包含在线人数）
type RoomWithOnlineCount struct {
	mysql.Room
	OnlineCount int `json:"OnlineCount"`
}

// RoomMember 房间成员信息
type RoomMember struct {
	UserID    string `json:"user_id"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Email     string `json:"email"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
	JoinedAt  int64  `json:"joined_at"`
}
