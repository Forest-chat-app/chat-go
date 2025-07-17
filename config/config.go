package config

type AppConfig struct {
	Server        Server         `mapstructure:"server" yaml:"server"`
	Mysql         Mysql          `mapstructure:"mysql" yaml:"mysql"`
	Mongo         Mongo          `mapstructure:"mongo" yaml:"mongo"`
	ElasticSearch ElasticSearch  `mapstructure:"elasticsearch" yaml:"elasticsearch"`
	Redis         Redis          `mapstructure:"redis" yaml:"redis"`
	Logger        Logger         `mapstructure:"logger" yaml:"logger"`
	MongoEsSync   []MongoEsSync  `mapstructure:"mongo_es_sync" yaml:"mongo_es_sync"`
	DBSchema      DBSchemaConfig `mapstructure:"db_schema" yaml:"db_schema"` // 新增字段
	JWT           JWT            `mapstructure:"jwt" yaml:"jwt"`             // JWT配置
}
