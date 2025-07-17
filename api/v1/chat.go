package v1

import (
	"chat-server/global"
	"chat-server/middleware"
	"chat-server/model/common"
	"chat-server/service"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type ChatApi struct{}

// WebSocketHandler 处理WebSocket连接
// @Summary 建立WebSocket连接
// @Description 建立WebSocket连接以接收和发送实时消息
// @Tags 聊天
// @Accept json
// @Produce json
// @Param room_id query string true "房间ID"
// @Security BearerAuth
// @Success 101 {string} string "Switching Protocols to WebSocket"
// @Router /api/v1/chat/ws [get]
func (chatApi *ChatApi) WebSocketHandler(c *gin.Context) {
	// 获取要连接的房间id
	roomId := c.Query("room_id")
	if roomId == "" {
		common.Result(c, common.INVALID_PARAMS)
		return
	}
	// 获取userId
	claims, exists := c.Get("claims")
	if !exists {
		common.Result(c, common.USER_NOT_FOUND)
		return
	}
	userId := claims.(*jwt.Token).Claims.(*middleware.AccessToken).UserID
	// 升级websocket连接
	conn, err := global.CHAT_UPGRADER.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		global.CHAT_LOG.Error("WebSocketHandler 升级websocket连接失败", "err", err)
		common.Result(c, common.ERROR)
		return
	}
	global.CHAT_LOG.Info("WebSocketHandler 升级websocket连接成功")
	// 创建客户端
	client := &service.Client{
		Conn:     conn,
		UserId:   userId,
		RoomId:   roomId,
		Send:     make(chan *service.WebSocketMessage, 256),
		LastPing: time.Now(),
		Manager:  global.CHAT_WEBSOCKET_MANAGER.(*service.WebSocketManager),
	}
	// 注册客户端
	client.Manager.Register <- client
	// 启动读取协程
	go client.ReadPump()
	go client.WritePump()

	//common.Result(c, common.SUCCESS, userId)
}
