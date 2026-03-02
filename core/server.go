package core

import (
	"chat-server/global"
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"
)

// RunServe 负责启动应用程序并处理关闭逻辑
func RunServe(appCtx context.Context, appCancel context.CancelFunc) {
	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", global.CHAT_CONFIG.Server.Host, global.CHAT_CONFIG.Server.Port),
		Handler: global.CHAT_ROUTERS,
	}

	// 启动服务器
	go func() {
		defer func() {
			if err := recover(); err != nil {
				global.CHAT_LOG.Error("HTTP服务器启动发生 panic", "stack_trace", string(debug.Stack()))
				appCancel()
			}
		}()
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			global.CHAT_LOG.Error("HTTP服务器启动失败", "err", err)
			appCancel()
		}
		global.CHAT_LOG.Info("HTTP服务器启动", "addr", srv.Addr)
	}()

	// 监听关闭信号
	select {
	case <-appCtx.Done():
		appCancel()
		global.CHAT_LOG.Info("应用程序 Context 已被取消，开始关闭...", "context_err", appCtx.Err())
	}

	// 关闭HTTP服务器
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		global.CHAT_LOG.Error("HTTP服务器关闭失败", "err", err)
	} else {
		global.CHAT_LOG.Info("HTTP服务器已关闭")
	}
	global.CHAT_LOG.Info("应用程序已关闭。")
}
