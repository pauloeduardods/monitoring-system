package monitoring_use_cases

import (
	"monitoring-system/src/internal/modules/monitoring/domain/camera"
	"monitoring-system/src/pkg/app_error"
	"monitoring-system/src/pkg/logger"
)

type CameraInfoUseCase interface {
	GetCameraDetails() ([]camera.CameraDetails, error)
}

type cameraInfoUseCase struct {
	cameraManager CameraManager
	logger        logger.Logger
}

func NewCameraInfoUseCase(cameraManager CameraManager, logger logger.Logger) CameraInfoUseCase {
	return &cameraInfoUseCase{
		cameraManager: cameraManager,
		logger:        logger,
	}
}

func (uc *cameraInfoUseCase) GetCameraDetails() ([]camera.CameraDetails, error) {
	cameras := uc.cameraManager.GetCameras()
	var cameraDetails []camera.CameraDetails

	for _, cam := range cameras {
		details := cam.GetDetails()
		cameraDetails = append(cameraDetails, details)
	}

	if len(cameraDetails) == 0 {
		uc.logger.Warning("No cameras detected")
		return nil, app_error.NewApiError(404, "No cameras detected")
	}

	return cameraDetails, nil
}
