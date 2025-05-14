package monitoring_use_cases

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"monitoring-system/src/config"
	"monitoring-system/src/internal/modules/monitoring/domain/camera"
	camera_infra "monitoring-system/src/internal/modules/monitoring/infra/camera"
	"monitoring-system/src/pkg/logger"
	"os"
	"runtime"
	"strings"
)

const DARWIN_MAX_CAMERAS = 3

type CameraManager interface {
	CheckSystemCameras() error
	GetCameras() map[string]camera.CameraService
	Close() error
}

type command struct {
	action func() error
	result chan error
}

type cameraManager struct {
	cameras     map[string]camera.CameraService
	logger      logger.Logger
	ctx         context.Context
	cancel      context.CancelFunc
	commandChan chan command
	config      *config.CameraConfig
}

func NewCameraManager(ctx context.Context, logger logger.Logger, config *config.CameraConfig) (CameraManager, error) {
	ctx, cancel := context.WithCancel(ctx)
	cm := &cameraManager{
		cameras:     make(map[string]camera.CameraService),
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
		commandChan: make(chan command),
		config:      config,
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

func (cm *cameraManager) newWebcam(deviceId interface{}) error {
	var id string
	switch v := deviceId.(type) {
	case int:
		id = fmt.Sprintf("%d", v)
	case string:
		id = fmt.Sprintf("%x", md5.Sum([]byte(v)))
	default:
		cm.logger.Error("deviceID is not of type int or string")
		return nil
	}

	if _, exists := cm.cameras[id]; exists {
		return errors.New("camera already exists")
	}

	webcam := camera_infra.NewCameraService(cm.ctx, id, deviceId, cm.logger, cm.config)

	err := webcam.Start()
	if err != nil {
		return err
	}

	cm.cameras[id] = webcam

	go func(id string) {
		select {
		case <-cm.ctx.Done():
		case <-webcam.Done():
			cm.execute(func() error {
				cm.logger.Info("Camera %d disconnected", id)
				delete(cm.cameras, id)
				return nil
			})
		}
	}(id)

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
			err := cm.newWebcam(i)
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
				err := cm.newWebcam(deviceID)
				if err != nil {
					continue
				}
			}
		}
		return nil
	})
}

func (cm *cameraManager) connectStreamCamera() error {
	cm.logger.Info("Connecting to stream camera")

	for _, stream := range cm.config.Stream {
		err := cm.newWebcam(stream.URL)
		if err != nil {
			cm.logger.Error("Error connecting to stream camera %s: %v", stream.URL, err)
			continue
		}
	}
	return nil
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

	err = cm.connectStreamCamera()
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

func (cm *cameraManager) GetCameras() map[string]camera.CameraService {
	return cm.cameras
}
