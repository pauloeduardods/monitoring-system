package config

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type AwsConfig struct {
	Region       string
	S3BucketName string
}

type ApiConfig struct {
	Host string
	Port int
}

type CamerasConfig struct {
	MaxCameraCount int
}

type Config struct {
	Aws     AwsConfig
	Api     ApiConfig
	Cameras CamerasConfig
}

type ConfigManager struct {
	db *sql.DB
}

func NewConfigManager(configDb *sql.DB) (*ConfigManager, error) {
	createAWSConfigTable := `
	CREATE TABLE IF NOT EXISTS aws_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		region TEXT NOT NULL,
		s3_bucket_name TEXT NOT NULL
	);`

	createApiConfigTable := `
	CREATE TABLE IF NOT EXISTS api_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		host TEXT NOT NULL,
		port INTEGER NOT NULL
	);`

	createCamerasConfigTable := `
	CREATE TABLE IF NOT EXISTS cameras_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		max_camera_count INTEGER NOT NULL
	);`

	_, err := configDb.Exec(createAWSConfigTable)
	if err != nil {
		return nil, err
	}

	_, err = configDb.Exec(createApiConfigTable)
	if err != nil {
		return nil, err
	}

	_, err = configDb.Exec(createCamerasConfigTable)
	if err != nil {
		return nil, err
	}

	err = insertDefaultConfigs(configDb)
	if err != nil {
		return nil, err
	}

	return &ConfigManager{db: configDb}, nil
}

func (cm *ConfigManager) LoadConfig() (*Config, error) {
	awsConfig, err := cm.loadAWSConfig()
	if err != nil {
		return nil, err
	}

	apiConfig, err := cm.loadApiConfig()
	if err != nil {
		return nil, err
	}

	camerasConfig, err := cm.loadCamerasConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		Aws:     *awsConfig,
		Api:     *apiConfig,
		Cameras: *camerasConfig,
	}, nil
}

func (cm *ConfigManager) loadAWSConfig() (*AwsConfig, error) {
	row := cm.db.QueryRow("SELECT region, s3_bucket_name FROM aws_config LIMIT 1")
	var region, s3BucketName string
	if err := row.Scan(&region, &s3BucketName); err != nil {
		return nil, fmt.Errorf("error loading AWS config: %v", err)
	}
	return &AwsConfig{Region: region, S3BucketName: s3BucketName}, nil
}

func (cm *ConfigManager) loadApiConfig() (*ApiConfig, error) {
	row := cm.db.QueryRow("SELECT host, port FROM api_config LIMIT 1")
	var host string
	var port int
	if err := row.Scan(&host, &port); err != nil {
		return nil, fmt.Errorf("error loading API config: %v", err)
	}
	return &ApiConfig{Host: host, Port: port}, nil
}

func (cm *ConfigManager) loadCamerasConfig() (*CamerasConfig, error) {
	row := cm.db.QueryRow("SELECT max_camera_count FROM cameras_config LIMIT 1")
	var maxCameraCount int
	if err := row.Scan(&maxCameraCount); err != nil {
		return nil, fmt.Errorf("error loading cameras config: %v", err)
	}

	return &CamerasConfig{MaxCameraCount: maxCameraCount}, nil
}

func (cm *ConfigManager) SaveAWSConfig(config *AwsConfig) error {
	_, err := cm.db.Exec("INSERT INTO aws_config (region, s3_bucket_name) VALUES (?, ?)", config.Region, config.S3BucketName)
	return err
}

func (cm *ConfigManager) SaveApiConfig(config *ApiConfig) error {
	_, err := cm.db.Exec("INSERT INTO api_config (host, port) VALUES (?, ?)", config.Host, config.Port)
	return err
}

func insertDefaultConfigs(db *sql.DB) error {
	row := db.QueryRow("SELECT COUNT(*) FROM aws_config")
	var count int
	if err := row.Scan(&count); err != nil {
		return fmt.Errorf("error checking AWS config table: %v", err)
	}
	if count == 0 {
		_, err := db.Exec("INSERT INTO aws_config (region, s3_bucket_name) VALUES (?, ?)", "us-east-1", "golang-camera")
		if err != nil {
			return fmt.Errorf("error inserting default AWS config: %v", err)
		}
	}

	row = db.QueryRow("SELECT COUNT(*) FROM api_config")
	if err := row.Scan(&count); err != nil {
		return fmt.Errorf("error checking API config table: %v", err)
	}
	if count == 0 {
		_, err := db.Exec("INSERT INTO api_config (host, port) VALUES (?, ?)", "0.0.0.0", 4000)
		if err != nil {
			return fmt.Errorf("error inserting default API config: %v", err)
		}
	}

	row = db.QueryRow("SELECT COUNT(*) FROM cameras_config")
	if err := row.Scan(&count); err != nil {
		return fmt.Errorf("error checking cameras config table: %v", err)
	}

	if count == 0 {
		_, err := db.Exec("INSERT INTO cameras_config (max_camera_count) VALUES (?)", 2)
		if err != nil {
			return fmt.Errorf("error inserting default cameras config: %v", err)
		}
	}

	return nil
}
