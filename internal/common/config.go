package common

import (
	"fmt"

	"github.com/spf13/viper"
)

// type AppConfig struct {
// 	App AppSettings `mapstructure:"app"`
// 	DB  DBConfig    `mapstructure:"db"`
// }

type AppConfig struct {
	App      AppSettings    `mapstructure:"app"`
	MongoDB  MongoDBConfig  `mapstructure:"mongodb"`
	Redis    RedisConfig    `mapstructure:"redis"`
	PortaOne PortaOneConfig `mapstructure:"portaone"`
}

type AppSettings struct {
	Env            string `mapstructure:"env"`
	GinMode        string `mapstructure:"gin_mode"`
	ServerAddress  string `mapstructure:"server_address"`
	SECRET_KEY     string `mapstructure:"secret_key"`
	S3_BUCKET_NAME string `mapstructure:"s3_bucket_name"`
}

type PortaOneConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// type DBConfig struct {
// 	URL string `mapstructure:"url"`
// }

type MongoDBConfig struct {
	URI      string `mapstructure:"url"`
	Database string `mapstructure:"database"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func LoadConfig() (*AppConfig, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config AppConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
