package routes

import (
	"monitoring-system/cmd/server/gin_server/handlers"

	"github.com/gin-gonic/gin"
)

func ConfigAuthRoutes(g *gin.RouterGroup, h *handlers.AuthHandler) {
	authGroup := g.Group("/auth")

	authGroup.POST("/login", h.Login())
	authGroup.POST("/register", h.Register())
}
