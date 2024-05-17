package main

import (
	"context"
	"database/sql"
	"fmt"
	"monitoring-system/cmd/api"
	"monitoring-system/cmd/websocket"
	"monitoring-system/config"
	"monitoring-system/internal/domain/camera"
	"monitoring-system/internal/storage"
	"monitoring-system/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Application struct {
	logger  logger.Logger
	storage storage.Storage
	config  *config.Config
	ctx     context.Context
	cam     camera.Camera
	wg      *sync.WaitGroup
	sqlDB   *sql.DB
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

	db, err := sql.Open("sqlite3", "./monitoring.db")
	if err != nil {
		logger.Error("Error opening database %v", err)
		return
	}
	defer db.Close()

	cam := camera.NewWebcam(ctx, 0, logger) //make multiple

	var wg sync.WaitGroup

	app := &Application{
		logger:  logger,
		storage: storage,
		config:  appConfig,
		ctx:     ctx,
		cam:     cam,
		wg:      &wg,
	}

	api := api.New(ctx, awsConfig, appConfig, logger, db)

	wg.Add(3)
	go func() {
		defer wg.Done()
		if err := api.Start(); err != nil {
			logger.Error("Error starting server %v", err)
		}
	}()
	go app.runApplication()
	go app.startServer()

	wg.Wait()
	<-ctx.Done()
	os.Exit(0)
}

func (a *Application) runApplication() {
	defer a.wg.Done()
	if err := a.cam.Start(); err != nil {
		a.logger.Error("Error starting camera %v", err)
		return
	}
	defer a.cam.Stop() //Check if this is the right place to put this

	filename := fmt.Sprintf("video_%s.avi", time.Now().Format("20060102_150405"))

	if err := a.cam.RecordVideo(a.ctx, filename); err != nil {
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

func (a *Application) startServer() {
	defer a.wg.Done()
	server := &http.Server{Addr: ":8080"}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			a.logger.Error("Error starting server %v", err)
		}
	}()

	wsServer := websocket.NewWebSocketServer(a.logger)
	http.HandleFunc("/video", func(w http.ResponseWriter, r *http.Request) {
		wsServer.VideoHandler(a.ctx, w, r, a.cam)
	})

	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/", fs)

	a.logger.Info("Server started at http://localhost:%d\n", 8080)

	<-a.ctx.Done()
	if err := server.Shutdown(context.Background()); err != nil {
		a.logger.Error("Error shutting down server %v", err)
	}
}
