package config

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

type Config struct {
	AWSRegion    string
	S3BucketName string
	DeviceID     []int
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

	deviceID := []int{0}

	return &Config{
		AWSRegion:    region,
		S3BucketName: s3BucketName,
		DeviceID:     deviceID,
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
