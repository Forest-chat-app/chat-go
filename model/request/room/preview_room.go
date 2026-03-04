package room

type PreviewRoomRequest struct {
	RoomId string `json:"room_id" binding:"required"`
}

type LeavePreviewRequest struct {
	RoomId string `json:"room_id" binding:"required"`
}
