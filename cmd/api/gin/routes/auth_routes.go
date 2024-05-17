package routes

import (
	"monitoring-system/cmd/api/gin/handlers"

	"github.com/gin-gonic/gin"
)

func ConfigAuthRoutes(g *gin.Engine, h *handlers.AuthHandler) {
	authGroup := g.Group("/api/v1/auth")

	authGroup.POST("/login", h.Login())
	authGroup.POST("/register", h.Register())
}
