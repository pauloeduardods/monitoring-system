package routes

import (
	"monitoring-system/src/api/gin_server/handlers"
	"monitoring-system/src/api/gin_server/middleware"

	"github.com/gin-gonic/gin"
)

func ConfigAuthRoutes(g *gin.RouterGroup, h *handlers.AuthHandler, m middleware.AuthMiddleware) {
	authGroup := g.Group("/auth")

	authGroup.POST("/login", h.Login())
	authGroup.POST("/register", m.AuthMiddlewareRegister(), h.Register())
}
