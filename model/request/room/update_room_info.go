package room

type UpdateRoomInfoRequest struct {
	RoomId       string `json:"room_id" binding:"required"`
	RoomName     string `json:"room_name"`
	Introduction string `json:"introduction"`
	Tag          string `json:"tag"`
	Status       int16  `json:"status"`
}
