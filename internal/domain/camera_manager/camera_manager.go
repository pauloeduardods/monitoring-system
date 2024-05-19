package camera_manager

import (
	"context"
	"errors"
	"monitoring-system/internal/domain/camera"
	"monitoring-system/pkg/logger"
	"sync"
)

const MAX_DEVICES = 10

type CameraManager interface {
	UpdateCameraStatus() error
	ListCameras() []Camera
	ListRunningCameras() []Camera
	GetCamera(deviceID int) (Camera, error)
	StartCamera(deviceID int) error
	StopCamera(deviceID int) error
}

type Status string

const (
	Connected    Status = "connected"
	Disconnected Status = "disconnected"
	Running      Status = "running"
)

type Camera struct {
	Status Status
	Id     int
	Camera camera.Camera
}

type cameraManager struct {
	mu      sync.Mutex
	cameras map[int]Camera
	logger  logger.Logger
	ctx     context.Context
	// sqlDB         *sql.DB //TODO: Save camera data to a database
}

func NewCameraManager(ctx context.Context, logger logger.Logger) CameraManager {
	cameras := make(map[int]Camera)
	for i := 0; i < MAX_DEVICES; i++ {
		cameras[i] = Camera{
			Status: Disconnected,
			Id:     i,
			Camera: camera.NewWebcam(ctx, i, logger),
		}
	}
	return &cameraManager{
		cameras: cameras,
		logger:  logger,
		ctx:     ctx,
	}
}

func (cm *cameraManager) UpdateCameraStatus() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	connectedCameras := cm.cameras
	for i := 0; i < MAX_DEVICES; i++ {
		if entry, ok := connectedCameras[i]; ok {
			_, err := entry.Camera.Check()
			if err != nil {
				entry.Status = Disconnected
				cm.cameras[i] = entry
				continue
			}
			entry.Status = Connected
			cm.cameras[i] = entry
		} else {
			cm.cameras[i] = Camera{
				Status: Disconnected,
				Id:     i,
				Camera: camera.NewWebcam(cm.ctx, i, cm.logger),
			}
		}
	}

	return nil
}

func (cm *cameraManager) ListCameras() []Camera {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cameras := make([]Camera, MAX_DEVICES)

	for i := 0; i < MAX_DEVICES; i++ {
		cameras[i] = cm.cameras[i]
	}

	return cameras
}

func (cm *cameraManager) ListRunningCameras() []Camera {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	runningCameras := make([]Camera, 0)

	for i := 0; i < MAX_DEVICES; i++ {
		if cm.cameras[i].Status == Running {
			runningCameras = append(runningCameras, cm.cameras[i])
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

func (cm *cameraManager) StartCamera(deviceID int) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cam, exists := cm.cameras[deviceID]
	if !exists {
		return errors.New("camera not found")
	}

	if cam.Status == Running {
		return nil
	}
	// if cam.Status == Disconnected {
	// 	return errors.New("camera is disconnected")
	// }

	return cam.Camera.Start()
}

func (cm *cameraManager) StopCamera(deviceID int) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cam, exists := cm.cameras[deviceID]
	if !exists {
		return errors.New("camera not found")
	}

	if cam.Status != Running {
		return nil
	}

	return cam.Camera.Stop()
}
