package config

type Logger struct {
	Level      string `mapstructure:"level" yaml:"level"`
	Output     string `mapstructure:"output" yaml:"output"`
	Format     string `mapstructure:"format" yaml:"format"`
	SourcePath bool   `mapstructure:"source_path" yaml:"source_path"`
}
