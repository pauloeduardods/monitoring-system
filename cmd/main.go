package main

import (
	"context"
	"fmt"
	"monitoring-system/config"
	"monitoring-system/internal/domain/camera"
	"monitoring-system/internal/domain/storage"
	"monitoring-system/pkg/logger"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Application struct {
	logger  *logger.Logger
	storage storage.Storage
	config  *config.Config
	ctx     context.Context
}

func main() {
	logger := logger.NewLogger()
	appConfig, err := config.NewConfig()
	if err != nil {
		logger.Error("Error loading configuration %v", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		logger.Info("Received signal: %v", sig)
		cancel()
	}()

	awsConfig, err := config.NewAWSConfig(ctx, appConfig)
	if err != nil {
		logger.Error("Error loading AWS configuration %v", err)
		return
	}

	storage, err := storage.NewStorage(logger, awsConfig, appConfig.S3BucketName)
	if err != nil {
		logger.Error("Error creating storage %v", err)
		return
	}

	app := &Application{
		logger:  logger,
		storage: storage,
		config:  appConfig,
		ctx:     ctx,
	}

	app.runApplication()
}

func (a *Application) runApplication() {
	cam := camera.NewWebcam(0, a.logger) //make multiple
	if err := cam.Start(); err != nil {
		a.logger.Error("Error starting camera %v", err)
		return
	}
	defer cam.Stop()

	filename := fmt.Sprintf("video_%s.avi", time.Now().Format("20060102_150405"))

	if err := cam.RecordVideo(a.ctx, filename); err != nil {
		a.logger.Error("Error recording video %v", err)
		return
	}
	a.logger.Info("Video recorded successfully")
	data, err := os.ReadFile(filename)
	if err != nil {
		a.logger.Error("Error reading video file %v", err)
		return
	}
	a.logger.Info("Uploading video to S3")

	if err := a.storage.Save(filename, data); err != nil {
		a.logger.Error("Error uploading video to S3 %v", err)
	} else {
		a.logger.Info("Video uploaded successfully")
	}
}
