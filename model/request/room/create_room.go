package room

type CreateRoom struct {
	RoomName     string `json:"room_name" binding:"required"`
	Introduction string `json:"introduction" binding:"required"`
	Tag          string `json:"tag" binding:"required"`
	Status       int16  `json:"status" binding:"required"`
}
