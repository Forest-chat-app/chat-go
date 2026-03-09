package chat

import "chat-server/model"

type SearchChatRequest struct {
	Content string `form:"content" binding:"required"`
	Type    int    `form:"type" binding:"required,oneof=1 2"`
}

type SearchChatResponse struct {
	RoomIds  []string             `json:"room_ids,omitempty"`
	Messages []model.UserMessages `json:"messages,omitempty"`
}
