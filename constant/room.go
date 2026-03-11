package constant

const (
	StatusPublic      = 1
	StatusPrivate     = 2
	StatusDelete      = 3
	StatusPublicName  = "公开"
	StatusPrivateName = "私密"
	StatusDeleteName  = "已删除"
	RoomAvatar        = "/avatar/defaultRoomAvatar.jpg"
)

var StatusMap = map[int16]bool{
	StatusPublic:  true,
	StatusPrivate: true,
	StatusDelete:  true,
}
