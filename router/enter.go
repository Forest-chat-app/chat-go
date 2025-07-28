package router

import (
	v1 "chat-server/api/v1"
)

var RouterGroupApp = new(RouterGroup)

type RouterGroup struct {
	UserRouter
	ChatRouter
	TokenRouter
	RoomRouter
}

var (
	userApi  = v1.ApiGroupApp.UserApi
	chatApi  = v1.ApiGroupApp.ChatApi
	tokenApi = v1.ApiGroupApp.TokenApi
	roomApi  = v1.ApiGroupApp.RoomApi
)
