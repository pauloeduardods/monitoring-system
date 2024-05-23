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

type Status string

const (
	Connected    Status = "connected"
	Disconnected Status = "disconnected"
	Removed      Status = "removed"
)

type CameraManager interface {
	CheckSystemCameras() error
	AddNotificationCallback(callback func(Camera)) error
	Close() error
}

type Camera struct {
	Status Status
	Camera camera.Camera
}

type command struct {
	action func() error
	result chan error
}

type cameraManager struct {
	cameras               map[int]*Camera
	logger                logger.Logger
	ctx                   context.Context
	cancel                context.CancelFunc
	notificationCallbacks []func(Camera)
	db                    *sql.DB
	commandChan           chan command
}

func NewCameraManager(ctx context.Context, logger logger.Logger, db *sql.DB) (CameraManager, error) {
	ctx, cancel := context.WithCancel(ctx)
	cm := &cameraManager{
		cameras:               make(map[int]*Camera),
		logger:                logger,
		ctx:                   ctx,
		cancel:                cancel,
		db:                    db,
		notificationCallbacks: []func(Camera){},
		commandChan:           make(chan command),
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

func (cm *cameraManager) AddNotificationCallback(callback func(Camera)) error {
	cm.logger.Info("Adding notification callback")
	return cm.execute(func() error {
		cm.logger.Info("Adding notification callback")
		cm.notificationCallbacks = append(cm.notificationCallbacks, callback)
		return nil
	})
}

func (cm *cameraManager) notifyStatusChange(cam Camera) {
	for _, callback := range cm.notificationCallbacks {
		callback(cam)
	}
}

func (cm *cameraManager) newWebcam(deviceId int) error {
	// return cm.execute(func() error {
	if currentCam, exists := cm.cameras[deviceId]; exists {
		if currentCam.Status == Connected {
			return nil
		}
		if currentCam.Status == Removed {
			return errors.New("camera is removed")
		}
	}

	webcam := camera.NewWebcam(cm.ctx, deviceId, cm.logger)
	cam := &Camera{
		Status: Disconnected,
		Camera: webcam,
	}

	cm.cameras[deviceId] = cam

	go func(deviceId int) {
		for status := range cm.cameras[deviceId].Camera.StatusChan() {
			_ = cm.execute(func() error {
				cam, exists := cm.cameras[deviceId]
				if !exists {
					cam = &Camera{
						Status: Disconnected,
						Camera: webcam,
					}
					cm.cameras[deviceId] = cam
				}

				switch status {
				case camera.Connected:
					if cam.Status == Connected || cam.Status == Removed {
						return nil
					}
					cam.Status = Connected
				case camera.Disconnected:
					if cam.Status == Disconnected || cam.Status == Removed {
						return nil
					}
					cam.Status = Disconnected
				}
				cm.notifyStatusChange(*cam)
				return nil
			})
		}
	}(deviceId)
	return nil
	// })
}

func (cm *cameraManager) startCameraNoLock(deviceID int) error {
	cam, exists := cm.cameras[deviceID]
	if !exists {
		err := cm.newWebcam(deviceID)
		if err != nil {
			return err
		}
		cam = cm.cameras[deviceID]
	}
	switch cam.Status {
	case Connected:
		return nil
	case Removed:
		return errors.New("camera is removed")
	case Disconnected:
	}
	return cam.Camera.Start()
	// })
}

func getDeviceID(deviceName string) int {
	var id int
	fmt.Sscanf(deviceName, "video%d", &id)
	return id
}

func (cm *cameraManager) checkMacCameras() error {
	return cm.execute(func() error {
		for i := 0; i < DARWIN_MAX_CAMERAS; i++ {
			err := cm.startCameraNoLock(i)
			if err != nil {
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
				err := cm.startCameraNoLock(deviceID)
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

	// go func() {
	// 	ticker := time.NewTicker(20 * time.Second)
	// 	defer ticker.Stop()
	// 	for {
	// 		select {
	// 		case <-cm.ctx.Done():
	// 			return
	// 		case <-ticker.C:
	// 			if err := cm.CheckSystemCameras(); err != nil {
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
			if cam.Status == Connected {
				err := cam.Camera.Close()
				if err != nil {
					cm.logger.Error("Error stopping camera %d: %v\n", i, err)
				}
			}
		}
		cm.cancel()
		close(cm.commandChan)
		return nil
	})
}
