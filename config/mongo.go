package config

type Mongo struct {
	URI             string `mapstructure:"uri" yaml:"uri"`
	DBName          string `mapstructure:"dbname" yaml:"dbname"`
	MinPoolSize     uint64 `mapstructure:"min_pool_size" yaml:"min_pool_size"`
	MaxPoolSize     uint64 `mapstructure:"max_pool_size" yaml:"max_pool_size"`
	ConnectTimeout  int    `mapstructure:"connect_timeout" yaml:"connect_timeout"`
	MaxConnIdleTime int    `mapstructure:"max_conn_idle_time" yaml:"max_conn_idle_time"`
}
