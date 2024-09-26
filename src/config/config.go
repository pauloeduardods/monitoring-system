package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type ApiConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type Config struct {
	Api    ApiConfig `mapstructure:"api"`
	JwtKey string    `mapstructure:"jwt_key"`
}

func setDefaults() {
	viper.SetDefault("api.host", "0.0.0.0")
	viper.SetDefault("api.port", 4000)
	viper.SetDefault("jwt_key", "SET_ME")
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
