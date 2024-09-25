package routes

import (
	"monitoring-system/src/api/gin_server/handlers"
	"monitoring-system/src/api/gin_server/middleware"

	"github.com/gin-gonic/gin"
)

func ConfigMonitoringRoutes(g *gin.RouterGroup, h *handlers.CameraHandler, m middleware.AuthMiddleware) {
	authGroup := g.Group("/monitoring")

	authGroup.GET("/camera/details", m.AuthMiddleware(), h.GetCameraDetails())
}
