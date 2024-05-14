package main

import (
	"context"
	"fmt"
	"monitoring-system/config"
	"monitoring-system/internal/domain/camera"
	"monitoring-system/internal/domain/storage"
	"monitoring-system/pkg/logger"
	"os"
	"time"
)

func main() {
	logger := logger.NewLogger()
	appConfig, err := config.NewConfig()
	if err != nil {
		logger.Error("Error loading configuration %v", err)
		return
	}

	awsConfig, err := config.NewAWSConfig(context.Background(), appConfig)
	if err != nil {
		logger.Error("Error loading AWS configuration %v", err)
		return
	}

	cam := camera.NewWebcam(appConfig.DeviceID[0])
	if err := cam.Start(); err != nil {
		logger.Error("Error starting camera %v", err)
		return
	}
	defer cam.Stop()

	duration := 5 * time.Second
	filename := fmt.Sprintf("video_%s.avi", time.Now().Format("20060102_150405"))
	if err := camera.RecordVideo(cam, filename, duration); err != nil {
		logger.Error("Error recording video %v", err)
		return
	}

	s3Storage, err := storage.NewStorage(logger, awsConfig, appConfig.S3BucketName)
	if err != nil {
		logger.Error("Error creating S3 storage %v", err)
		return
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		logger.Error("Error reading video file %v", err)
		return
	}

	if err := s3Storage.Save(filename, data); err != nil {
		logger.Error("Error uploading video to S3 %v", err)
	} else {
		logger.Info("Video uploaded successfully")
	}
}
