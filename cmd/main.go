package main

import (
	"context"
	"database/sql"
	"fmt"
	"monitoring-system/cmd/modules"
	"monitoring-system/cmd/server"
	"monitoring-system/config"
	"monitoring-system/internal/domain/camera_manager"
	"monitoring-system/pkg/logger"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
)

// type Application struct {
// 	logger  logger.Logger
// 	storage storage.Storage
// 	config  *config.Config
// 	ctx     context.Context
// 	cm      camera_manager.CameraManager
// 	sqlDB   *sql.DB
// 	modules *modules.Modules
// }

func main() {
	logger, err := logger.NewLogger("development")
	if err != nil {
		fmt.Printf("Error creating logger %v", err)
		return
	}

	db, err := sql.Open("sqlite3", "./monitoring.db")
	if err != nil {
		logger.Error("Error opening database %v", err)
		return
	}
	defer db.Close()

	configManager, err := config.NewConfigManager(db)
	if err != nil {
		logger.Error("Error creating config manager %v", err)
		return
	}

	appConfig, err := configManager.LoadConfig()
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

	awsConfig, err := config.LoadAwsConfig(ctx, appConfig.Aws, logger)
	if err != nil {
		logger.Error("Error loading AWS configuration %v", err)
		return
	}

	// storage, err := storage.NewStorage(logger, awsConfig, appConfig.Aws.S3BucketName)
	// if err != nil {
	// 	logger.Error("Error creating storage %v", err)
	// 	return
	// }

	cm, err := camera_manager.NewCameraManager(ctx, logger)
	if err != nil {
		logger.Error("Error creating camera manager %v", err)
		return
	}

	// cam := camera.NewWebcam(ctx, 0, logger)
	// err = cam.Start()
	// if err != nil {
	// 	logger.Error("Error starting camera %v", err)
	// 	return
	// }
	// err = cam.RecordVideo(ctx, "video.avi")
	// if err != nil {
	// 	logger.Error("Error recording video %v", err)
	// 	return
	// }

	modules, err := modules.New(logger, db, cm)
	if err != nil {
		logger.Error("Error creating modules %v", err)
		return
	}

	// app := &Application{
	// 	logger:  logger,
	// 	storage: storage,
	// 	config:  appConfig,
	// 	ctx:     ctx,
	// 	cm:      cm,
	// 	sqlDB:   db,
	// 	modules: modules,
	// }

	server := server.New(ctx, awsConfig, appConfig, logger, modules)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.Start(); err != nil {
			logger.Error("Error starting server %v", err)
		}
	}()
	// go func() {
	// 	defer wg.Done()
	// 	logger.Info("Starting application")
	// 	app.runApplication()
	// }()

	if err := cm.CheckSystemCameras(); err != nil {
		logger.Error("Error updating camera status %v", err)
		return
	}

	wg.Wait()

	<-ctx.Done()
	os.Exit(0)
}

// func (a *Application) runApplication() {

// 	filename := fmt.Sprintf("video_%s.avi", time.Now().Format("20060102_150405"))

// 	cam, err := a.cm.GetCamera(0)
// 	if err != nil {
// 		a.logger.Error("Error getting camera %v", err)
// 		return
// 	}

// 	if err := cam.Camera.RecordVideo(a.ctx, filename); err != nil {
// 		a.logger.Error("Error recording video %v", err)
// 		return
// 	}
// 	a.logger.Info("Video recorded successfully")
// 	data, err := os.ReadFile(filename)
// 	if err != nil {
// 		a.logger.Error("Error reading video file %v", err)
// 		return
// 	}
// 	a.logger.Info("Uploading video to S3")

// 	if err := a.storage.Save(filename, data); err != nil {
// 		a.logger.Error("Error uploading video to S3 %v", err)
// 	} else {
// 		a.logger.Info("Video uploaded successfully")
// 	}
// }
