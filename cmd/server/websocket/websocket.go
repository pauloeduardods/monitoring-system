package websocket

import (
	"context"
	"monitoring-system/cmd/modules"
	"monitoring-system/cmd/server/gin_server/middleware"
	"monitoring-system/cmd/server/websocket/handler"
	"monitoring-system/pkg/logger"

	"github.com/gin-gonic/gin"
)

type WebSocketServer struct {
	logger         logger.Logger
	gin            *gin.RouterGroup
	ctx            context.Context
	modules        *modules.Modules
	authMiddleware middleware.AuthMiddleware
}

func NewWebSocketServer(ctx context.Context, logger logger.Logger, gin *gin.RouterGroup, mod *modules.Modules, authMiddleware middleware.AuthMiddleware) *WebSocketServer {
	return &WebSocketServer{
		logger:         logger,
		gin:            gin,
		ctx:            ctx,
		modules:        mod,
		authMiddleware: authMiddleware,
	}

}
func (wss *WebSocketServer) Start() {
	wss.logger.Info("Starting websocket server")
	cam, err := wss.modules.Internal.CameraManager.GetCamera(0)
	if err != nil {
		wss.logger.Error("Error getting camera %v", err)
		return
	}

	authMiddleware := wss.authMiddleware.AuthMiddlewareWS()

	handler := handler.NewVideoHandler(wss.ctx, cam.Camera, wss.logger)
	wss.gin.GET("/video", authMiddleware, func(c *gin.Context) {
		handler.VideoHandler(c.Writer, c.Request)
	})
}
