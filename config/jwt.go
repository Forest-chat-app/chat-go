package config

type JWT struct {
	Secret         string `mapstructure:"secret" yaml:"secret"`             // JWT密钥
	AccessTime     int    `mapstructure:"access_time" yaml:"access_time"`   // 过期时间（天）
	RefreshTime    int    `mapstructure:"refresh_time" yaml:"refresh_time"` // 过期时间（天）
	UserTokensTime int    `mapstructure:"user_tokens_time" yaml:"user_tokens_time"`
	Issuer         string `mapstructure:"issuer" yaml:"issuer"` // 签发人
}
