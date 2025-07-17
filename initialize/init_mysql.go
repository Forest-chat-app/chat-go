package initialize

import (
	"chat-server/global"
	"context"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"time"
)

func InitMySQL() error {
	global.CHAT_LOG.Info("初始化mysql")
	// 获取反序列化后的mysql
	mysqlConfig := global.CHAT_CONFIG.Mysql
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		mysqlConfig.User, mysqlConfig.Password, mysqlConfig.Host, mysqlConfig.Port, mysqlConfig.DBName, mysqlConfig.Charset)

	// 连接mysql
	var err error
	global.CHAT_MYSQL, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 输出所有 SQL
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
	})
	if err != nil {
		global.CHAT_LOG.Error("mysql连接失败：", "err", err)
		return err
	}
	// 获取底层的 *sql.DB 对象
	sqlDB, err := global.CHAT_MYSQL.DB()
	if err != nil {
		global.CHAT_LOG.Error("获取 GORM 底层 *sql.DB 失败: ", "err", err)
		return err
	}
	// 设置连接池参数
	sqlDB.SetConnMaxLifetime(time.Duration(mysqlConfig.MaxLifetime) * time.Second) // 连接最大生命周期
	sqlDB.SetMaxOpenConns(mysqlConfig.MaxOpenConns)                                // 最大打开连接数
	sqlDB.SetMaxIdleConns(mysqlConfig.MaxIdleConns)

	// 连接超时用于 PingContext，确保连接创建和测试在规定时间内完成
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(mysqlConfig.ConnTimeout)*time.Second)
	defer cancel()

	if err = sqlDB.PingContext(ctx); err != nil {
		global.CHAT_LOG.Error("Ping MySQL 失败: ", "err", err)
		closeErr := sqlDB.Close()
		if closeErr != nil {
			global.CHAT_LOG.Error("Ping MySQL 失败后，关闭 MySQL 失败: ", "err", closeErr)
		}
		return err
	}
	global.CHAT_LOG.Info("mysql连接成功")
	return nil
}
