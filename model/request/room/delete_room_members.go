package room

type DeleteRoomMembersRequest struct {
	RoomId  string   `json:"room_id" binding:"required"`
	UserIds []string `json:"user_ids" binding:"required,min=1"`
}
