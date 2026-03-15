package room

type UpdateRoomAvatarRequest struct {
	RoomId string `json:"room_id" binding:"required"`
	Avatar string `json:"avatar" binding:"required"`
}
