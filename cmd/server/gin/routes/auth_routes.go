package routes

import (
	"monitoring-system/cmd/server/gin/handlers"

	"github.com/gin-gonic/gin"
)

func ConfigAuthRoutes(g *gin.Engine, h *handlers.AuthHandler) {
	authGroup := g.Group("/api/v1/auth")

	authGroup.POST("/login", h.Login())
	authGroup.POST("/register", h.Register())
}
