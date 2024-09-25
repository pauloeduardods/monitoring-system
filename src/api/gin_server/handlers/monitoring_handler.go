package handlers

import (
	monitoring_use_cases "monitoring-system/src/internal/modules/monitoring/usecases"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CameraHandler struct {
	uc *monitoring_use_cases.MonitoringUseCases
}

func NewCameraHandler(uc *monitoring_use_cases.MonitoringUseCases) *CameraHandler {
	return &CameraHandler{
		uc: uc,
	}
}

func (a *CameraHandler) GetCameraDetails() gin.HandlerFunc {
	return func(g *gin.Context) {
		res, err := a.uc.CameraInfoUseCase.GetCameraDetails()
		if err != nil {
			g.Error(err)
			return
		} else {
			g.JSON(http.StatusOK, res)
		}
	}
}
