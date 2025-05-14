package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type ApiConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type StreamConfig struct {
	URL        string `mapstructure:"url"`
	StreamName string `mapstructure:"stream_name"`
}

type CameraConfig struct {
	FPS                int            `mapstructure:"fps"`
	Width              int            `mapstructure:"width"`
	Height             int            `mapstructure:"height"`
	Codec              string         `mapstructure:"codec"`
	MotionDetection    bool           `mapstructure:"motion_detection"`
	MinArea            int            `mapstructure:"min_area"`
	CheckSystemCameras bool           `mapstructure:"check_system_cameras"`
	Stream             []StreamConfig `mapstructure:"stream"`
}

type Config struct {
	Api    ApiConfig    `mapstructure:"api"`
	JwtKey string       `mapstructure:"jwt_key"`
	Camera CameraConfig `mapstructure:"camera"`
}

func setDefaults() {
	viper.SetDefault("api.host", "0.0.0.0")
	viper.SetDefault("api.port", 4000)
	viper.SetDefault("jwt_key", "SET_ME")
	viper.SetDefault("camera", CameraConfig{
		FPS:                15,
		Width:              640,
		Height:             480,
		Codec:              "MJPG",
		MotionDetection:    true,
		MinArea:            4000,
		CheckSystemCameras: true,
		Stream:             []StreamConfig{},
	})

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
