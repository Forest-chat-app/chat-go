package room

type GetRoomInfoRequest struct {
	RoomId string `form:"room_id" binding:"required"`
}
