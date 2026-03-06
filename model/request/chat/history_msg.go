package chat

import "chat-server/model"

type HistoryMsgRequest struct {
	RoomId   string `form:"room_id" binding:"required"`
	PageNum  int    `form:"page_num" binding:"required"`
	PageSize int    `form:"page_size" binding:"required"`
}

type HistoryMsgResponse struct {
	Messages []model.UserMessages `json:"messages"`
}
