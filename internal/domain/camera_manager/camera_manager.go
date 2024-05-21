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

	_ "github.com/mattn/go-sqlite3"
)

const DARWIN_MAX_CAMERAS = 20

type CameraManager interface {
	CheckSystemCameras() error
	ListCameras() []Camera
	ListRunningCameras() []Camera
	GetCamera(deviceID int) (Camera, error)
	AddCamera(deviceID int, name string) error
	RemoveCamera(deviceID int) error
	RenameCamera(deviceID int, name string) error
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
	mu      sync.Mutex
	cameras map[int]Camera
	logger  logger.Logger
	ctx     context.Context
	db      *sql.DB
}

func NewCameraManager(ctx context.Context, logger logger.Logger, db *sql.DB) (CameraManager, error) {
	cm := &cameraManager{
		cameras: make(map[int]Camera),
		logger:  logger,
		ctx:     ctx,
		db:      db,
	}

	err := cm.loadCamerasFromDB()
	if err != nil {
		return nil, err
	}

	return cm, nil
}

func (cm *cameraManager) startCameraNoLock(deviceID int) error {
	cam, exists := cm.cameras[deviceID]
	if !exists {
		cam = Camera{
			Id:     deviceID,
			Name:   "",
			Status: Disconnected,
			Camera: camera.NewWebcam(cm.ctx, deviceID, cm.logger),
		}
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
	cm.cameras[deviceID] = cam
	cm.saveCameraToDB(cam)
	return nil
}

func getDeviceID(deviceName string) int {
	var id int
	fmt.Sscanf(deviceName, "video%d", &id)
	return id
}

func (cm *cameraManager) CheckSystemCameras() error {
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
	var cam Camera
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
		cam = Camera{
			Id:     deviceID,
			Name:   "",
			Status: Connected,
			Camera: webCam,
		}
		_, err := cam.Camera.Check()
		if err != nil {
			cam.Status = Disconnected
		}
	}
	cm.logger.Info("Camera %d status: %s", deviceID, cam.Status)
	if cam.Status == Connected {
		err := cm.startCameraNoLock(deviceID)
		if err != nil {
			cm.logger.Error("Error starting camera %d: %v\n", deviceID, err)
		}
	}
	cm.cameras[deviceID] = cam
	cm.saveCameraToDB(cam)
}

func (cm *cameraManager) ListCameras() []Camera {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cameras := make([]Camera, 0)
	for _, cam := range cm.cameras {
		cameras = append(cameras, cam)
	}

	return cameras
}

func (cm *cameraManager) ListRunningCameras() []Camera {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	runningCameras := make([]Camera, 0)

	for _, cam := range cm.cameras {
		if cam.Status == Running {
			runningCameras = append(runningCameras, cam)
		}
	}

	return runningCameras
}

func (cm *cameraManager) GetCamera(deviceID int) (Camera, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.logger.Info("Getting camera %d", deviceID)
	cam, exists := cm.cameras[deviceID]
	if !exists {
		cam = Camera{
			Id:     deviceID,
			Name:   "",
			Status: Disconnected,
			Camera: camera.NewWebcam(cm.ctx, deviceID, cm.logger),
		}
	}

	return cam, nil
}

func (cm *cameraManager) AddCamera(deviceID int, name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cam, exists := cm.cameras[deviceID]
	if !exists {
		cam = Camera{
			Id:     deviceID,
			Name:   "",
			Status: Disconnected,
			Camera: camera.NewWebcam(cm.ctx, deviceID, cm.logger),
		}
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
	cm.cameras[deviceID] = cam
	return cm.saveCameraToDB(cam)
}

func (cm *cameraManager) RemoveCamera(deviceID int) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	entry, exists := cm.cameras[deviceID]
	if !exists {
		entry = Camera{
			Id:     deviceID,
			Name:   "",
			Status: Disconnected,
			Camera: camera.NewWebcam(cm.ctx, deviceID, cm.logger),
		}
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
	return cm.saveCameraToDB(cm.cameras[deviceID])
}

func (cm *cameraManager) RenameCamera(deviceID int, name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cam, exists := cm.cameras[deviceID]
	if !exists {
		cam = Camera{
			Id:     deviceID,
			Name:   "",
			Status: Disconnected,
			Camera: camera.NewWebcam(cm.ctx, deviceID, cm.logger),
		}
	}

	cam.Name = name
	cm.cameras[deviceID] = cam
	return cm.saveCameraToDB(cam)
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
