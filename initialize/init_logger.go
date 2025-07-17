package initialize

import (
	"chat-server/global"
	"fmt"
	"log/slog"
	"os"
)

func InitLogger() error {
	fmt.Println("开始初始化日志功能")
	loggerConfig := global.CHAT_CONFIG.Logger

	// 1、设置日志级别
	var level slog.Level
	switch loggerConfig.Level {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// 2、设置日志处理器
	opts := &slog.HandlerOptions{
		AddSource: loggerConfig.SourcePath,
		Level:     level,
	}
	var handler slog.Handler
	switch loggerConfig.Format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts) // JSON 格式输出到标准输出
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts) // Text 格式输出到标准输出
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts) // 默认 JSON
	}

	// 3、设置全局默认 Logger
	global.CHAT_LOG = slog.New(handler)
	slog.SetDefault(global.CHAT_LOG)

	fmt.Printf("日志配置已加载：级别=%s, 输出=%s, 格式=%s, 源文件=%t\n",
		loggerConfig.Level, loggerConfig.Output, loggerConfig.Format, loggerConfig.SourcePath)

	return nil
}
