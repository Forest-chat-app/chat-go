package v1

import "chat-server/service"

var ApiGroupApp = new(ApiGroup)

type ApiGroup struct {
	UserApi
	ChatApi
	TokenApi
	RoomApi
	CdnApi
}

var (
	chatService  = service.ServiceGroupApp.ChatService
	userService  = service.ServiceGroupApp.UserService
	tokenService = service.ServiceGroupApp.TokenService
	roomService  = service.ServiceGroupApp.RoomService
	cdnService   = service.ServiceGroupApp.CdnService
)
