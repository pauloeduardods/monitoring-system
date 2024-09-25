package monitoring_use_cases

import "monitoring-system/src/pkg/logger"

type MonitoringUseCases struct {
	CameraInfoUseCase CameraInfoUseCase
}

func NewMonitoringUseCases(logger logger.Logger, cm CameraManager) *MonitoringUseCases {
	return &MonitoringUseCases{
		CameraInfoUseCase: NewCameraInfoUseCase(cm, logger),
	}
}
