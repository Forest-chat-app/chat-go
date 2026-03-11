package initialize

import (
	"chat-server/global"
	"fmt"
)

func InitCdn() error {
	global.CHAT_LOG.Info("正在同步 CDN/OSS 配置状态...")
	cfg := global.CHAT_CONFIG.CdnConfig

	// 1. 校验基础 AK 信息
	if cfg.AccessKeyId == "" || cfg.AccessKeySecret == "" {
		return fmt.Errorf("CDN 初始化失败: 缺失 AccessKeyId 或 AccessKeySecret")
	}

	// 2. 校验角色信息（因为你用了 STS 模式，这是必须的）
	if cfg.RoleArn == "" {
		return fmt.Errorf("CDN 初始化失败: 缺失 RoleArn (STS模式核心参数)")
	}

	// 3. 默认值兜底
	if cfg.SessionName == "" {
		cfg.SessionName = "ChatApp_Default_Session"
	}
	if cfg.TokenExpire == 0 {
		cfg.TokenExpire = 3600 // 默认一小时
	}

	global.CHAT_LOG.Info(fmt.Sprintf("CDN 初始化成功: 区域[%s], 桶[%s]", cfg.Region, cfg.BucketName))
	return nil
}
