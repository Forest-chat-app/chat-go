package common

// WebSocket消息结构
type WebSocketMessage struct {
	Type      string      `json:"type"`
	RoomId    string      `json:"room_id"`
	SenderId  string      `json:"sender_id"`
	Content   interface{} `json:"content"`
	CreatedAt int64       `json:"created_at"`
}
