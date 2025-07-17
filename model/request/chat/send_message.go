package chat

// 发送消息请求结构
type SendMessageRequest struct {
	Content string `json:"content" binding:"required"`
	ToUser  string `json:"to_user" binding:"required"`
}
