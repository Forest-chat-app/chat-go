package room

type QuitRoomRequest struct {
	RoomId string `json:"room_id" binding:"required"`
}
