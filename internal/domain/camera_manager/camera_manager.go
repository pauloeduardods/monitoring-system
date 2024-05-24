package camera_manager

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"monitoring-system/internal/domain/camera"
	"monitoring-system/pkg/logger"
	"os"
	"runtime"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const DARWIN_MAX_CAMERAS = 3

type CameraManager interface {
	CheckSystemCameras() error
	GetCameras() map[int]camera.Camera
	// AddNotificationCallback(callback func(camera.Camera)) error
	Close() error
}

type command struct {
	action func() error
	result chan error
}

type cameraManager struct {
	cameras     map[int]camera.Camera
	logger      logger.Logger
	ctx         context.Context
	cancel      context.CancelFunc
	db          *sql.DB
	commandChan chan command
}

func NewCameraManager(ctx context.Context, logger logger.Logger, db *sql.DB) (CameraManager, error) {
	ctx, cancel := context.WithCancel(ctx)
	cm := &cameraManager{
		cameras:     make(map[int]camera.Camera),
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
		db:          db,
		commandChan: make(chan command),
	}

	go cm.run()

	return cm, nil
}

func (cm *cameraManager) run() {
	for cmd := range cm.commandChan {
		err := cmd.action()
		cmd.result <- err
		close(cmd.result)
	}
	cm.logger.Warning("Camera manager command channel closed")
}

func (cm *cameraManager) execute(action func() error) error {
	cmd := command{action: action, result: make(chan error)}
	cm.commandChan <- cmd
	return <-cmd.result
}

func (cm *cameraManager) newWebcam(deviceId int) error {
	// return cm.execute(func() error {
	if _, exists := cm.cameras[deviceId]; exists {
		return errors.New("camera already exists")
	}

	webcam := camera.NewWebcam(cm.ctx, deviceId, cm.logger)

	err := webcam.Start()
	if err != nil {
		return err
	}

	cm.cameras[deviceId] = webcam

	go func(deviceId int) {
		select {
		case <-cm.ctx.Done():
		case <-webcam.Done():
			cm.execute(func() error {
				cm.logger.Info("Camera %d disconnected", deviceId)
				delete(cm.cameras, deviceId)
				return nil
			})
		}
	}(deviceId)

	return nil
}

func getDeviceID(deviceName string) int {
	var id int
	fmt.Sscanf(deviceName, "video%d", &id)
	return id
}

func (cm *cameraManager) checkMacCameras() error {
	return cm.execute(func() error {
		for i := 0; i < DARWIN_MAX_CAMERAS; i++ {
			if _, exists := cm.cameras[i]; exists {
				continue
			}
			err := cm.newWebcam(i)
			if err != nil {
				cm.logger.Error("Error checking camera %d: %v", i, err)
				continue
			}

		}
		return nil
	})
}

func (cm *cameraManager) checkLinuxCameras() error {
	return cm.execute(func() error {
		devices, err := os.ReadDir("/dev")
		if err != nil {
			return err
		}

		for _, device := range devices {
			deviceName := device.Name()
			if strings.HasPrefix(deviceName, "video") {
				deviceID := getDeviceID(deviceName)
				if _, exists := cm.cameras[deviceID]; exists {
					continue
				}
				err := cm.newWebcam(deviceID)
				if err != nil {
					cm.logger.Error("Error checking camera %d: %v", deviceID, err)
					continue
				}
			}
		}
		return nil
	})
}

func (cm *cameraManager) checkSystemCameras() error {
	cm.logger.Info("Checking system cameras")

	switch runtime.GOOS {
	case "linux":
		return cm.checkLinuxCameras()
	case "darwin":
		return cm.checkMacCameras()
	default:
		return errors.New("unsupported operating system")
	}
}

func (cm *cameraManager) CheckSystemCameras() error {
	err := cm.checkSystemCameras()
	if err != nil {
		return err
	}

	// go func() { //TODO: Fix camera reconnection
	// 	ticker := time.NewTicker(20 * time.Second)
	// 	defer ticker.Stop()
	// 	for {
	// 		select {
	// 		case <-cm.ctx.Done():
	// 			return
	// 		case <-ticker.C:
	// 			if err := cm.checkSystemCameras(); err != nil {
	// 				cm.logger.Error("Error updating camera status %v", err)
	// 			}
	// 		}
	// 	}
	// }()

	return nil
}

func (cm *cameraManager) Close() error {
	return cm.execute(func() error {
		for i, cam := range cm.cameras {
			err := cam.Close()
			if err != nil {
				cm.logger.Error("Error stopping camera %d: %v\n", i, err)
			}
		}
		cm.cancel()
		close(cm.commandChan)
		return nil
	})
}

func (cm *cameraManager) GetCameras() map[int]camera.Camera {
	return cm.cameras
}
