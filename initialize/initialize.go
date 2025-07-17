package initialize

import (
	"chat-server/global"
	"chat-server/service"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync" // 导入 sync 包
)

// Initialize 函数负责所有应用程序的初始化
// 它接受一个 context 和 WaitGroup，用于统一管理 Goroutine 生命周期
func Initialize(appCtx context.Context, appCancel context.CancelFunc, wg *sync.WaitGroup) error { // 接受 appCtx 和 wg

	// 初始化配置文件 (检查错误)
	if err := InitConfig(); err != nil {
		return fmt.Errorf("初始化配置文件失败: %w", err)
	}

	// 初始化日志 (检查错误)
	if err := InitLogger(); err != nil {
		return fmt.Errorf("初始化日志工具失败: %w", err)
	}

	// 初始化数据库 (检查错误)
	if err := InitMySQL(); err != nil {
		return fmt.Errorf("初始化 MySQL 失败: %w", err)
	}
	if err := InitRedis(); err != nil {
		return fmt.Errorf("初始化 Redis 失败: %w", err)
	}
	if err := InitMongo(); err != nil {
		return fmt.Errorf("初始化 MongoDB 失败: %w", err)
	}
	if err := InitElasticSearch(); err != nil {
		return fmt.Errorf("初始化 Elasticsearch 失败: %w", err)
	}

	// 数据库结构检测与创建 (传递 AppConfig.DBSchema)
	if err := InitDatabaseSchemas(appCtx, global.CHAT_CONFIG.DBSchema); err != nil { // <-- 修改这里
		return fmt.Errorf("初始化数据库结构失败: %w", err)
	}

	// 启动数据同步服务 (或其他后台服务)
	if len(global.CHAT_CONFIG.MongoEsSync) == 0 {
		slog.Warn("未配置任何 MongoDB 到 Elasticsearch 的同步对，跳过启动数据同步服务。")
	} else {
		slog.Info("开始为配置的同步对启动数据同步服务...")
		for _, pair := range global.CHAT_CONFIG.MongoEsSync {
			p := pair
			wg.Add(1) // 增加 Goroutine 计数
			go func() {
				defer wg.Done() // Goroutine 结束时递减计数
				defer func() {
					if r := recover(); r != nil {
						slog.Error(fmt.Sprintf("Mongo-ES 同步 Goroutine (Collection: %s, Index: %s) 发生 panic", p.MongoCollection, p.EsIndex),
							"panic_value", r, "stack_trace", string(debug.Stack()))
						appCancel() // 如果 Goroutine 发生 panic，立即触发全局取消
					}
				}()
				slog.Info(fmt.Sprintf("启动 Mongo-to-ES 数据同步服务 (Collection: %s -> Index: %s)...", p.MongoCollection, p.EsIndex))
				syncErr := service.ServiceGroupApp.StartMongoToEsSync(appCtx, p.MongoCollection, p.EsIndex)
				if syncErr != nil && !errors.Is(syncErr, context.Canceled) {
					slog.Error(fmt.Sprintf("MongoDB 到 Elasticsearch 同步服务 (Collection: %s -> Index: %s) 终止，发生非取消错误", p.MongoCollection, p.EsIndex), "err", syncErr)
					appCancel()
				} else if errors.Is(syncErr, context.Canceled) {
					slog.Info(fmt.Sprintf("MongoDB 到 Elasticsearch 同步服务 (Collection: %s -> Index: %s) 已因 Context 取消而停止。", p.MongoCollection, p.EsIndex))
				}
			}()
		}
	}

	// 初始化路由
	InitRouter()

	// 开启WebSocket
	StartWebSocketManager(appCtx)

	slog.Info("所有应用程序组件初始化完成。")
	return nil
}
