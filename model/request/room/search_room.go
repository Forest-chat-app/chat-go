package room

type SearchRoomRequest struct {
	Content string `form:"content" binding:"required"`
	Type    int16  `form:"type" binding:"required"`
}
