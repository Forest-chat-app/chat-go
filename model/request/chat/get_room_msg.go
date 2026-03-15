package chat

import "chat-server/model"

type GetRoomMsgReq struct {
	RoomId  string `form:"room_id" binding:"required"`
	Content string `form:"content" binding:"required"`
}

type GetRoomMsgRsp struct {
	Messages []model.UserMessages `json:"messages,omitempty"`
}
