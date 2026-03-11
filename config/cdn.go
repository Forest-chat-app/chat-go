package config

type CdnConfig struct {
	AccessKeyId     string `mapstructure:"access_key_id" json:"access_key_id" yaml:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret" json:"access_key_secret" yaml:"access_key_secret"`
	Region          string `mapstructure:"region" json:"region" yaml:"region"`
	BucketName      string `mapstructure:"bucket_name" json:"bucket_name" yaml:"bucket_name"`
	OssHost         string `mapstructure:"oss_host" json:"oss_host" yaml:"oss_host"`
	RoleArn         string `mapstructure:"role_arn" json:"role_arn" yaml:"role_arn"`
	SessionName     string `mapstructure:"session_name" json:"session_name" yaml:"session_name"`
	UploadDir       string `mapstructure:"upload_dir" json:"upload_dir" yaml:"upload_dir"`
	TokenExpire     int    `mapstructure:"token_expire" json:"token_expire" yaml:"token_expire"` // 签名有效期（秒）
	Domain          string `mapstructure:"domain" json:"domain" yaml:"domain"`
	PrivateKey      string `mapstructure:"private_key" json:"private_key" yaml:"private_key"`
	ExpireDuration  int64  `mapstructure:"expire_duration" json:"expire_duration" yaml:"expire_duration"` // 签名有效期（秒）
}
