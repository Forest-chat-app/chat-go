package config

type DBSchemaConfig struct {
	MySQL         *MySQLSchemaConfig                `mapstructure:"mysql" yaml:"mysql"`
	MongoDB       *MongoDBClusterSchemaConfig       `mapstructure:"mongodb" yaml:"mongodb"`
	Elasticsearch *ElasticsearchClusterSchemaConfig `mapstructure:"elasticsearch" yaml:"elasticsearch"`
}

// MySQLSchemaConfig 定义MySQL Schema配置
type MySQLSchemaConfig struct {
	ScriptFile string `mapstructure:"script_file" yaml:"script_file"` // 指向SQL脚本文件
}

// MongoDBCollectionSchema 定义单个MongoDB集合的Schema配置
type MongoDBCollectionSchema struct {
	Name                 string `mapstructure:"name" yaml:"name"`
	IndexFile            string `mapstructure:"index_file" yaml:"index_file"`
	ValidatorCommandFile string `mapstructure:"validator_command_file" yaml:"validator_command_file"` // 新增字段
}

// MongoDBClusterSchemaConfig 定义整个MongoDB集群的Schema配置
type MongoDBClusterSchemaConfig struct {
	Collections []MongoDBCollectionSchema `mapstructure:"collections" yaml:"collections"`
}

// MongoDBIndexSchema 定义单个MongoDB索引的配置 (用于解析JSON文件中的索引)
type MongoDBIndexSchema struct {
	Keys    map[string]interface{} `json:"keys"`
	Options map[string]interface{} `mapstructure:"options" yaml:"options"`
}

// ElasticsearchIndexSchema 定义单个Elasticsearch索引的配置
type ElasticsearchIndexSchema struct {
	Name        string `mapstructure:"name" yaml:"name"`
	RequestFile string `mapstructure:"request_file" yaml:"request_file"` // 指向完整的JSON请求体文件
}

// ElasticsearchClusterSchemaConfig 定义整个Elasticsearch集群的Schema配置
type ElasticsearchClusterSchemaConfig struct {
	Indices []ElasticsearchIndexSchema `mapstructure:"indices" yaml:"indices"`
}
