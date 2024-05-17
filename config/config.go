package config

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

type Config struct {
	AWSRegion    string
	S3BucketName string
	DeviceID     []int
	Host         string
	Port         int
}

func NewConfig() (*Config, error) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		return nil, fmt.Errorf("AWS_REGION environment variable is not set")
	}

	s3BucketName := os.Getenv("S3_BUCKET_NAME")
	if s3BucketName == "" {
		return nil, fmt.Errorf("S3_BUCKET_NAME environment variable is not set")
	}

	host := os.Getenv("HOST")
	if host == "" {
		return nil, fmt.Errorf("HOST environment variable is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		return nil, fmt.Errorf("PORT environment variable is not set")
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("error converting PORT to int: %v", err)
	}

	deviceID := []int{0}

	return &Config{
		AWSRegion:    region,
		S3BucketName: s3BucketName,
		DeviceID:     deviceID,
		Host:         host,
		Port:         portInt,
	}, nil
}

func NewAWSConfig(ctx context.Context, env *Config) (*aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(env.AWSRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("error loading aws configuration: %v", err)
	}

	return &cfg, nil
}
