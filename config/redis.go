package config

type Redis struct {
	Address      string `mapstructure:"address" yaml:"address"`
	Username     string `mapstructure:"username" yaml:"username"`
	Password     string `mapstructure:"password" yaml:"password"`
	DB           int    `mapstructure:"db" yaml:"db"`
	PoolSize     int    `mapstructure:"pool_size" yaml:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns" yaml:"min_idle_conns"`
	PoolTimeout  int    `mapstructure:"pool_timeout" yaml:"pool_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout" yaml:"idle_timeout"`
}
