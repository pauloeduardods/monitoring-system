package websocket

import (
	"context"
	"monitoring-system/cmd/modules"
	"monitoring-system/cmd/server/gin_server/middleware"
	"monitoring-system/cmd/server/websocket/handler"
	"monitoring-system/internal/domain/camera_manager"
	"monitoring-system/pkg/app_error"
	"monitoring-system/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type WebSocketServer struct {
	logger         logger.Logger
	gin            *gin.RouterGroup
	ctx            context.Context
	modules        *modules.Modules
	authMiddleware middleware.AuthMiddleware
	cameras        map[int]*camera_manager.Camera
}

func NewWebSocketServer(ctx context.Context, logger logger.Logger, gin *gin.RouterGroup, mod *modules.Modules, authMiddleware middleware.AuthMiddleware) *WebSocketServer {
	return &WebSocketServer{
		logger:         logger,
		gin:            gin,
		ctx:            ctx,
		modules:        mod,
		authMiddleware: authMiddleware,
		cameras:        make(map[int]*camera_manager.Camera),
	}
}

func (wss *WebSocketServer) notificationCallback(cam *camera_manager.Camera) {
	wss.logger.Info("Camera notification %v", cam)
	switch cam.Status {
	case camera_manager.Running:
		wss.cameras[cam.Id] = cam
	default:
		delete(wss.cameras, cam.Id)
	}
}

func (wss *WebSocketServer) videoHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Error(app_error.NewApiError(http.StatusBadRequest, "Invalid camera id"))
		return
	}
	cam, ok := wss.cameras[id]
	if !ok {
		c.Error(app_error.NewApiError(http.StatusNotFound, "Camera not found"))
		return
	}

	handler := handler.NewVideoHandler(wss.ctx, cam.Camera, wss.logger)
	handler.VideoHandler(c.Writer, c.Request)
}

func (wss *WebSocketServer) Start() error {
	wss.logger.Info("Starting websocket server")
	err := wss.modules.Internal.CameraManager.AddNotificationCallback(wss.notificationCallback)
	if err != nil {
		wss.logger.Error("Error adding notification callback %v", err)
		return err
	}

	authMiddleware := wss.authMiddleware.AuthMiddlewareWS()

	wss.gin.GET("/video/:id", authMiddleware, wss.videoHandler)

	return nil
}
