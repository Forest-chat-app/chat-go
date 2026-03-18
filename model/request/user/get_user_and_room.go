package user

type GetUserAndRoomReq struct {
	UserId string `form:"user_id" binding:"required"`
}
