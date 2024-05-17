package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"monitoring-system/pkg/logger"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3 struct {
	logger logger.Logger
	client *s3.Client
	bucket string
}

func NewStorage(logger logger.Logger, awsConf *aws.Config, bucket string) (*S3, error) {
	logger.Info("Creating new S3 storage")

	return &S3{
		logger: logger,
		client: s3.NewFromConfig(*awsConf),
		bucket: bucket,
	}, nil
}

func getContentTypes(extension string) string {
	switch extension {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "pdf":
		return "application/pdf"
	case "txt":
		return "text/plain"
	case "json":
		return "application/json"
	case "mp4":
		return "video/mp4"
	case "avi":
		return "video/x-msvideo"
	case "mkv":
		return "video/x-matroska"
	default:
		return "application/octet-stream"
	}
}

func (s *S3) Save(key string, data []byte) error {
	s.logger.Info("Uploading file to S3 %s with key: %s", s.bucket, key)
	keySplit := strings.Split(key, ".")
	extension := keySplit[len(keySplit)-1]

	_, err := s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(getContentTypes(extension)),
	})
	if err != nil {
		s.logger.Error("Error uploading file to S3 %s: %v", s.bucket, err)
		return fmt.Errorf("error uploading file to s3 %s: %v", s.bucket, err)
	}
	s.logger.Info("File uploaded to S3 %s with key: %s", s.bucket, key)
	return nil
}

func (s *S3) Get(key string) ([]byte, error) {
	s.logger.Info("Getting file from S3 %s with key: %s", s.bucket, key)
	result, err := s.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") {
			return nil, ErrFileNotFound
		}
		s.logger.Error("Error getting file from S3 %s Key: %s Err: %v", s.bucket, key, err)
		return nil, fmt.Errorf("error getting file from s3 %s key: %s err: %v", s.bucket, key, err)
	}
	defer result.Body.Close()
	data, err := io.ReadAll(result.Body)
	if err != nil {
		s.logger.Error("Error reading file from S3 %s Key: %s Err: %v", s.bucket, key, err)
		return nil, fmt.Errorf("error reading file from s3 %s key: %s err: %v", s.bucket, key, err)
	}

	return data, nil
}
