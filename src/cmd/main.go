package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	server "monitoring-system/src/api"
	"monitoring-system/src/config"
	"monitoring-system/src/factory"
	"monitoring-system/src/pkg/logger"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Println("Starting monitoring system...")
	configPath := flag.String("config", ".", "Path to the configuration file")
	saveData := flag.String("save-data", ".", "Path where program data will be saved")
	staticFiles := flag.String("static-files", "src/web/static", "Path to static files")

	flag.Parse()

	logger, err := logger.NewLogger("development")
	if err != nil {
		fmt.Printf("Error creating logger %v", err)
		return
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("%s/%s", *saveData, "monitoring.db"))
	if err != nil {
		logger.Error("Error opening database %v", err)
		return
	}
	defer db.Close()

	appConfig, err := config.LoadConfig(*configPath)
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

	factory, err := factory.NewFactory(ctx, logger, db, appConfig)
	if err != nil {
		logger.Error("Error creating factory %v", err)
		return
	}

	err = factory.Monitoring.CameraManager.CheckSystemCameras()
	if err != nil {
		logger.Error("Error Checking system cameras", err)
		return
	}

	server := server.New(appConfig, logger, factory)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.Start(ctx, *staticFiles); err != nil {
			logger.Error("Error starting server %v", err)
		}
	}()

	wg.Wait()

	<-ctx.Done()
	os.Exit(0)
}
