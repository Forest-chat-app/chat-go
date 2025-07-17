package config

type ElasticSearch struct {
	Address            string `mapstructure:"address" yaml:"address"`
	ApiKey             string `mapstructure:"api_key" yaml:"api_key"`
	Username           string `mapstructure:"username" yaml:"username"`
	Password           string `mapstructure:"password" yaml:"password"`
	InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify" yaml:"insecure_skip_verify"`
	CaFile             string `mapstructure:"ca_file" yaml:"ca_file"`
}
