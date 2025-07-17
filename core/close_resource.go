package core

import (
	"chat-server/global"
	"context"
	"time"
)

func CloseResource() {

	if global.CHAT_MYSQL != nil {
		global.CHAT_LOG.Info("关闭 MySQL 连接...")
		if sqlDB, err := global.CHAT_MYSQL.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				global.CHAT_LOG.Error("关闭 MySQL 连接失败:", "err", err)
			} else {
				global.CHAT_LOG.Info("MySQL 连接已关闭")
			}
		} else {
			global.CHAT_LOG.Error("获取 GORM 底层 *sql.DB 失败，无法关闭 MySQL:", "err", err)
		}
	}

	if global.CHAT_REDIS != nil {
		global.CHAT_LOG.Info("关闭 Redis 连接...")
		if err := global.CHAT_REDIS.Close(); err != nil {
			global.CHAT_LOG.Error("关闭 Redis 连接失败:", "err", err)
		} else {
			global.CHAT_LOG.Info("Redis 连接已关闭")
		}
	}

	if global.CHAT_MONGO != nil {
		global.CHAT_LOG.Info("关闭 MongoDB 连接...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := global.CHAT_MONGO.Disconnect(ctx); err != nil {
			global.CHAT_LOG.Error("关闭 MongoDB 连接失败:", "err", err)
		} else {
			global.CHAT_LOG.Info("MongoDB 连接已关闭")
		}
	}
}
