package room

type JoinRoomRequest struct {
	RoomId string `json:"room_id" binding:"required"`
}
