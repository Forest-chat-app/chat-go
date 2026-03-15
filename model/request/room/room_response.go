package room

import "chat-server/model/mysql"

// RoomWithOnlineCount 房间信息（包含在线人数）
type RoomWithOnlineCount struct {
	mysql.Room
	OnlineCount int `json:"OnlineCount"`
}

// RoomMember 房间成员信息
type RoomMember struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Email    string `json:"email"`
	JoinedAt int64  `json:"joined_at"`
}
