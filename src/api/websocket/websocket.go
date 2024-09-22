package websocket

import (
	"context"

	"monitoring-system/src/api/gin_server/middleware"
	"monitoring-system/src/api/websocket/handler"
	"monitoring-system/src/factory"
	"monitoring-system/src/pkg/app_error"
	"monitoring-system/src/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type WebSocketServer struct {
	logger         logger.Logger
	gin            *gin.RouterGroup
	ctx            context.Context
	factory        *factory.Factory
	authMiddleware middleware.AuthMiddleware
}

func NewWebSocketServer(ctx context.Context, logger logger.Logger, gin *gin.RouterGroup, fac *factory.Factory, authMiddleware middleware.AuthMiddleware) *WebSocketServer {
	return &WebSocketServer{
		logger:         logger,
		gin:            gin,
		ctx:            ctx,
		factory:        fac,
		authMiddleware: authMiddleware,
	}
}

func (wss *WebSocketServer) videoHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Error(app_error.NewApiError(http.StatusBadRequest, "Invalid camera id"))
		return
	}
	cam, ok := wss.factory.Monitoring.CameraManager.GetCameras()[id]
	if !ok {
		c.Error(app_error.NewApiError(http.StatusNotFound, "Camera not found"))
		return
	}

	handler := handler.NewVideoHandler(wss.ctx, cam, wss.logger)
	handler.VideoHandler(c.Writer, c.Request)
}

func (wss *WebSocketServer) Start() error {
	wss.logger.Info("Starting websocket server")

	wss.logger.Info("Added notification callback")

	authMiddleware := wss.authMiddleware.AuthMiddlewareWS()

	wss.gin.GET("/video/:id", authMiddleware, wss.videoHandler)

	return nil
}
