package config

type MongoEsSync struct {
	MongoCollection string `mapstructure:"mongo_collection" yaml:"mongo_collection"`
	EsIndex         string `mapstructure:"es_index" yaml:"es_index"`
}
