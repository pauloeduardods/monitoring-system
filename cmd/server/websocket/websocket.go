package websocket

import (
	"context"
	"monitoring-system/cmd/server/websocket/handler"
	"monitoring-system/internal/domain/camera"
	"monitoring-system/pkg/logger"

	"github.com/gin-gonic/gin"
)

type WebSocketServer struct {
	logger logger.Logger
	gin    *gin.RouterGroup
	ctx    context.Context
	cam    camera.Camera
}

func NewWebSocketServer(ctx context.Context, logger logger.Logger, gin *gin.RouterGroup, cam camera.Camera) *WebSocketServer {
	return &WebSocketServer{
		logger: logger,
		gin:    gin,
		ctx:    ctx,
		cam:    cam,
	}
}

func (wss *WebSocketServer) Start() {
	wss.logger.Info("Starting websocket server")
	handler := handler.NewVideoHandler(wss.ctx, wss.cam, wss.logger)
	wss.gin.GET("/video", func(c *gin.Context) {
		handler.VideoHandler(c.Writer, c.Request)
	})
}
