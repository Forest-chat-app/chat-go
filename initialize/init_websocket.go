package initialize

import (
	"chat-server/global"
	"chat-server/service"
	"context"
	"github.com/gorilla/websocket"
	"net/http"
)

// StartWebSocketManager 启动WebSocket管理器
func StartWebSocketManager(ctx context.Context) {
	global.CHAT_LOG.Info("开始启动WebSocket管理器")
	// 定义WebSocket升级器
	global.CHAT_UPGRADER = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有跨域请求，生产环境应该限制
		},
	}
	// 定义全局WebSocketManager
	global.CHAT_WEBSOCKET_MANAGER = service.NewWebSocketManager()
	// 启动WebSocket管理器
	go global.CHAT_WEBSOCKET_MANAGER.(*service.WebSocketManager).Run(ctx)
}
