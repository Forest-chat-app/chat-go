package v1

import (
	"chat-server/core"
	"chat-server/global"
	"chat-server/middleware"
	"chat-server/model/common"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
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
// @Router /api/v1/chat/webSocketHandler [get]
func (chatApi *ChatApi) WebSocketHandler(c *gin.Context) {
	// 1、获取userId
	var userId string
	claims, exists := c.Get("claims")
	if !exists {
		// 1.1 如果中间件没拿到，我们尝试从 URL 参数拿到 token 并手动解析一次
		token := c.Query("token")
		if token == "" {
			global.CHAT_LOG.Error("WebSocket 连接缺失 Token")
			return
		}
		// 1.2 调用你的 JWT 工具类手动解析
		claims, err := jwt.ParseWithClaims(token, &middleware.AccessToken{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(global.CHAT_CONFIG.JWT.Secret), nil
		})
		if err != nil || !claims.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"data":    nil,
				"message": "未授权",
			})
			c.Abort()
			return
		}
	} else {
		userId = claims.(*jwt.Token).Claims.(*middleware.AccessToken).UserID
	}

	// 2、升级websocket连接
	conn, err := global.CHAT_UPGRADER.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		global.CHAT_LOG.Error("WebSocketHandler 升级websocket连接失败", "err", err)
		common.Result(c, common.ERROR)
		return
	}
	global.CHAT_LOG.Info("WebSocketHandler 升级websocket连接成功")
	// 3、创建客户端
	client := &core.Client{
		Conn:     conn,
		UserId:   userId,
		Send:     make(chan *common.WebSocketMessage, 256),
		LastPing: time.Now(),
		Manager:  global.CHAT_WEBSOCKET_MANAGER.(*core.WebSocketManager),
	}
	// 3.1 上线客户端
	client.Manager.Register <- client
	// 3.2 客户端启动读写协程
	go client.ReadPump()
	go client.WritePump()
}
