package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type AwsConfig struct {
	Region       string `mapstructure:"region"`
	S3BucketName string `mapstructure:"s3_bucket_name"`
}

type ApiConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type Config struct {
	Aws AwsConfig `mapstructure:"aws"`
	Api ApiConfig `mapstructure:"api"`
}

func setDefaults() {
	viper.SetDefault("aws.region", "us-east-1")
	viper.SetDefault("aws.s3_bucket_name", "golang-camera")
	viper.SetDefault("api.host", "0.0.0.0")
	viper.SetDefault("api.port", 4000)
}

func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)

	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := viper.SafeWriteConfigAs(configPath + "/config.yaml"); err != nil {
				return nil, fmt.Errorf("error writing default config file: %v", err)
			}
		} else {
			return nil, fmt.Errorf("error reading config file: %v", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %v", err)
	}

	return &config, nil
}
