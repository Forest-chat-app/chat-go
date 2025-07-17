package config

type Mysql struct {
	Host         string `mapstructure:"host" yaml:"host"`
	Port         int    `mapstructure:"port" yaml:"port"`
	User         string `mapstructure:"user" yaml:"user"`
	Password     string `mapstructure:"password" yaml:"password"`
	DBName       string `mapstructure:"dbname" yaml:"dbname"`
	Charset      string `mapstructure:"charset" yaml:"charset"`
	MaxLifetime  int    `mapstructure:"max_lifetime" yaml:"max_lifetime"`
	MaxOpenConns int    `mapstructure:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns" yaml:"max_idle_conns"`
	ConnTimeout  int    `mapstructure:"conn_timeout" yaml:"conn_timeout"`
}
