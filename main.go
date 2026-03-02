package main

import (
	"chat-server/core"
	_ "chat-server/docs"
	"chat-server/initialize"
	"log/slog"
	"os"

	"context"
	"sync"
)

// @title           Chat Server API接口文档
// @version         1.0
// @description     聊天服务器API文档
// @termsOfService  http://swagger.io/terms/

// @contact.name   汤新康
// @contact.url    http://www.example.com/support
// @contact.email  2912528586@qq.com

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 请在此输入Bearer令牌，格式为: Bearer {token}
// @BasePath                    /
func main() {
	// 设置全局context，用来统一关闭goroutine
	appCtx, appCancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	defer appCancel()
	defer core.CloseResource()

	// 1、初始化各种设置
	err := initialize.Initialize(appCtx, appCancel, &wg)
	if err != nil {
		slog.Error("应用程序初始化失败:", "err", err)
		os.Exit(1)
	}

	// 2、启动服务
	core.RunServe(appCtx, appCancel)
}
