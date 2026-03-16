package room

type DeleteRoomRequest struct {
	RoomId string `json:"room_id" binding:"required"`
}
