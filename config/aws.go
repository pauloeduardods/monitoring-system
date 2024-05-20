package config

import (
	"context"
	"monitoring-system/pkg/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/smithy-go/logging"
)

func LoadAwsConfig(ctx context.Context, awsConfig AwsConfig, logger logger.Logger) (*aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsConfig.Region), config.WithLogger(createAWSLogAdapter(logger)))
	if err != nil {
		return nil, err
	}

	cfg.Region = awsConfig.Region

	return &cfg, nil
}

func createAWSLogAdapter(log logger.Logger) logging.LoggerFunc {
	return func(classification logging.Classification, format string, v ...interface{}) {
		switch classification {
		case logging.Debug:
			log.Debug(format, v...)
		case logging.Warn:
			log.Warning(format, v...)
		default:
			log.Info(format, v...)
		}
	}
}
