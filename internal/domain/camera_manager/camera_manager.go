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
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const DARWIN_MAX_CAMERAS = 20

type CameraManager interface {
	CheckSystemCameras() error
	AddCamera(deviceID int, name string) error
	RemoveCamera(deviceID int) error
	RenameCamera(deviceID int, name string) error
	AddNotificationCallback(callback func(*Camera)) error
	Close() error
}

type Status string

const (
	Connected    Status = "connected"
	Disconnected Status = "disconnected"
	Running      Status = "running"
	Removed      Status = "removed"
)

type Camera struct {
	Id     int
	Name   string
	Status Status
	Camera camera.Camera
}

type cameraManager struct {
	mu                    sync.Mutex
	cameras               map[int]*Camera
	logger                logger.Logger
	ctx                   context.Context
	db                    *sql.DB
	notificationCallbacks []func(*Camera)
}

func NewCameraManager(ctx context.Context, logger logger.Logger, db *sql.DB) (CameraManager, error) {
	cm := &cameraManager{
		cameras:               make(map[int]*Camera),
		logger:                logger,
		ctx:                   ctx,
		db:                    db,
		notificationCallbacks: []func(*Camera){},
	}

	return cm, nil
}

func (cm *cameraManager) AddNotificationCallback(callback func(*Camera)) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.notificationCallbacks = append(cm.notificationCallbacks, callback)
	return nil
}

func (cm *cameraManager) notifyStatusChange(cam *Camera) {
	for _, callback := range cm.notificationCallbacks {
		callback(cam)
	}
}

func (cm *cameraManager) startCameraNoLock(deviceID int) error {
	cam, exists := cm.cameras[deviceID]
	if !exists {
		cam = &Camera{
			Id:     deviceID,
			Name:   "",
			Status: Disconnected,
			Camera: camera.NewWebcam(cm.ctx, deviceID, cm.logger),
		}
		cm.cameras[deviceID] = cam
	}

	switch cam.Status {
	case Running:
		return nil
	case Removed:
		return errors.New("camera is removed")
	case Disconnected:
	case Connected:
	}

	if err := cam.Camera.Start(); err != nil {
		return err
	}
	cam.Status = Running
	cm.saveCameraToDB(*cam)
	cm.notifyStatusChange(cam)
	return nil
}

func getDeviceID(deviceName string) int {
	var id int
	fmt.Sscanf(deviceName, "video%d", &id)
	return id
}

func (cm *cameraManager) checkSystemCameras() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.logger.Info("Checking system cameras")

	if runtime.GOOS == "linux" {
		return cm.checkLinuxCameras()
	} else if runtime.GOOS == "darwin" {
		return cm.checkMacCameras()
	}

	return errors.New("unsupported operating system")
}

func (cm *cameraManager) CheckSystemCameras() error {
	err := cm.loadCamerasFromDB()
	if err != nil {
		return err
	}
	err = cm.checkSystemCameras()
	if err != nil {
		return err
	}

	go func() { //TODO: Test this
		ticker := time.NewTicker(20 * time.Second) //TODO: Check if this is the right interval
		defer ticker.Stop()
		for {
			select {
			case <-cm.ctx.Done():
				return
			case <-ticker.C:
				if err := cm.CheckSystemCameras(); err != nil {
					cm.logger.Error("Error updating camera status %v", err)
				}
			}
		}
	}()

	return nil
}

func (cm *cameraManager) checkLinuxCameras() error {
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
			webCam := camera.NewWebcam(cm.ctx, deviceID, cm.logger)
			_, err := webCam.Check()
			if err == nil {
				cm.checkAndStartCamera(deviceID)
			}
		}
	}

	return nil
}

func (cm *cameraManager) checkMacCameras() error {
	for i := 0; i < DARWIN_MAX_CAMERAS; i++ {
		if _, exists := cm.cameras[i]; exists {
			continue
		}
		webCam := camera.NewWebcam(cm.ctx, i, cm.logger)
		_, err := webCam.Check()
		if err == nil {
			cm.checkAndStartCamera(i)
		}
	}

	return nil
}

func (cm *cameraManager) checkAndStartCamera(deviceID int) {
	var cam *Camera
	if entry, ok := cm.cameras[deviceID]; ok {
		cam = entry
		_, err := entry.Camera.Check()
		if err != nil {
			cam.Status = Disconnected
		} else {
			cam.Status = Connected
		}
	} else {
		webCam := camera.NewWebcam(cm.ctx, deviceID, cm.logger)
		cam = &Camera{
			Id:     deviceID,
			Name:   "",
			Status: Connected,
			Camera: webCam,
		}
		_, err := cam.Camera.Check()
		if err != nil {
			cam.Status = Disconnected
		}
		cm.cameras[deviceID] = cam
	}
	cm.logger.Info("Camera %d status: %s", deviceID, cam.Status)
	if cam.Status == Connected {
		err := cm.startCameraNoLock(deviceID)
		if err != nil {
			cm.logger.Error("Error starting camera %d: %v\n", deviceID, err)
		}
	}
	cm.saveCameraToDB(*cam)
	cm.notifyStatusChange(cam)
}

func (cm *cameraManager) AddCamera(deviceID int, name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cam, exists := cm.cameras[deviceID]
	if !exists {
		cam = &Camera{
			Id:     deviceID,
			Name:   "",
			Status: Disconnected,
			Camera: camera.NewWebcam(cm.ctx, deviceID, cm.logger),
		}
		cm.cameras[deviceID] = cam
	}

	switch cam.Status {
	case Running:
		return errors.New("camera is running")
	case Disconnected:
		return errors.New("camera is disconnected")
	case Connected:
	case Removed:
	}
	cam.Name = name
	cam.Status = Connected
	cm.notifyStatusChange(cam)
	return cm.saveCameraToDB(*cam)
}

func (cm *cameraManager) RemoveCamera(deviceID int) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	entry, exists := cm.cameras[deviceID]
	if !exists {
		entry = &Camera{
			Id:     deviceID,
			Name:   "",
			Status: Disconnected,
			Camera: camera.NewWebcam(cm.ctx, deviceID, cm.logger),
		}
		cm.cameras[deviceID] = entry
	}

	switch entry.Status {
	case Disconnected:
		return errors.New("camera is disconnected")
	case Removed:
		return errors.New("camera is already removed")
	case Connected:
	case Running:
	}

	err := entry.Camera.Stop()
	if err != nil {
		return err
	}

	entry.Status = Removed
	cm.notifyStatusChange(entry)
	return cm.saveCameraToDB(*entry)
}

func (cm *cameraManager) RenameCamera(deviceID int, name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cam, exists := cm.cameras[deviceID]
	if !exists {
		cam = &Camera{
			Id:     deviceID,
			Name:   "",
			Status: Disconnected,
			Camera: camera.NewWebcam(cm.ctx, deviceID, cm.logger),
		}
		cm.cameras[deviceID] = cam
	}

	cam.Name = name
	cm.notifyStatusChange(cam)
	return cm.saveCameraToDB(*cam)
}

func (cm *cameraManager) Close() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, cam := range cm.cameras {
		if cam.Status == Running {
			err := cam.Camera.Stop()
			if err != nil {
				cm.logger.Error("Error stopping camera %d: %v\n", cam.Id, err)
			}
		}
	}

	return nil
}
