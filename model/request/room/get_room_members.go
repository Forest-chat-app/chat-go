package room

type GetRoomMembersRequest struct {
	RoomId string `form:"room_id" binding:"required"`
}
