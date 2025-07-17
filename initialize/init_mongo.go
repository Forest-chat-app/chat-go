package initialize

import (
	"chat-server/global"
	"context"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

func InitMongo() error {
	global.CHAT_LOG.Info("初始化mongo")
	mongoConfig := global.CHAT_CONFIG.Mongo
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	clientOpts := options.Client().ApplyURI(mongoConfig.URI).SetServerAPIOptions(serverAPI).
		SetMinPoolSize(mongoConfig.MinPoolSize).                                     // 最小连接数
		SetMaxPoolSize(mongoConfig.MaxPoolSize).                                     // 最大连接数
		SetConnectTimeout(time.Duration(mongoConfig.ConnectTimeout) * time.Second).  // 连接超时时间
		SetMaxConnIdleTime(time.Duration(mongoConfig.MaxConnIdleTime) * time.Second) // 连接最大空闲时间

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		global.CHAT_LOG.Error("连接 MongoDB 失败: ", "err", err)
		return err
	}

	// 检查连接
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(mongoConfig.ConnectTimeout)*time.Second)
	defer cancel()
	if err = client.Ping(ctx, nil); err != nil {
		global.CHAT_LOG.Error("Ping MongoDB 失败", "err", err)
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer disconnectCancel()
		closeErr := client.Disconnect(disconnectCtx)
		if closeErr != nil {
			global.CHAT_LOG.Error("Ping MongoDB 失败后，关闭 MongoDB 失败: ", "err", closeErr)
		}
		return err
	}

	global.CHAT_MONGO = client
	global.CHAT_MONGODB = client.Database(mongoConfig.DBName)
	global.CHAT_LOG.Info("MongoDB连接成功")
	return nil
}
