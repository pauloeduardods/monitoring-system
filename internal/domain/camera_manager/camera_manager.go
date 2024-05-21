package camera_manager

import (
	"context"
	"database/sql"
	"errors"
	"monitoring-system/config"
	"monitoring-system/internal/domain/camera"
	"monitoring-system/pkg/logger"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

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
	mu           sync.Mutex
	cameras      map[int]Camera
	logger       logger.Logger
	ctx          context.Context
	cameraConfig config.CamerasConfig
	db           *sql.DB
}

func NewCameraManager(ctx context.Context, logger logger.Logger, db *sql.DB, cameraConfig config.CamerasConfig) (CameraManager, error) {
	cm := &cameraManager{
		cameras:      make(map[int]Camera),
		logger:       logger,
		ctx:          ctx,
		db:           db,
		cameraConfig: cameraConfig,
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
		return errors.New("camera not found")
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

func (cm *cameraManager) CheckSystemCameras() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.logger.Info("Checking cameras")

	for i := 0; i < cm.cameraConfig.MaxCameraCount; i++ {
		var cam Camera
		if entry, ok := cm.cameras[i]; ok {
			cam = entry
			_, err := entry.Camera.Check()
			if err != nil {
				cam.Status = Disconnected
			} else {
				cam.Status = Connected
			}
		} else {
			var cam Camera
			webCam := camera.NewWebcam(cm.ctx, i, cm.logger)
			cam.Camera = webCam
			cam.Id = i
			cam.Name = ""
			_, err := cam.Camera.Check()
			if err != nil {
				cam.Status = Disconnected
			} else {
				cam.Status = Connected
			}
		}
		cm.logger.Info("Camera %d status: %s", i, cam.Status)
		if cam.Status == Connected {
			err := cm.startCameraNoLock(i)
			if err != nil {
				cm.logger.Error("Error starting camera %d: %v\n", i, err)
			}
		}
		cm.cameras[i] = cam
		cm.saveCameraToDB(cam)
	}

	return nil
}

func (cm *cameraManager) ListCameras() []Camera {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cameras := make([]Camera, 0, cm.cameraConfig.MaxCameraCount)
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
		return Camera{}, errors.New("camera not found")
	}

	return cam, nil
}

func (cm *cameraManager) AddCamera(deviceID int, name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cam, exists := cm.cameras[deviceID]
	if !exists {
		return errors.New("camera not found")
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
		return errors.New("camera not found")
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
		return errors.New("camera not found")
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
