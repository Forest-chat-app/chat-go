package initialize

import (
	"chat-server/global"
	"fmt"
	"github.com/spf13/viper"
	"log/slog"
)

func InitConfig() error {
	fmt.Println("初始化配置文件")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		slog.Error("viper 读取配置失败, err: ", "err", err)
		return err
	}
	if err := viper.Unmarshal(&global.CHAT_CONFIG); err != nil {
		slog.Error("viper 配置反序列化失败, err: ", "err", err)
		return err
	}
	fmt.Println("配置文件初始化成功")
	return nil
}
