package response

import "chat-server/model/mysql"

// RoomWithOnlineCount 房间信息（包含在线人数）
type RoomWithOnlineCount struct {
	mysql.Room
	OnlineCount int `json:"OnlineCount"`
}
