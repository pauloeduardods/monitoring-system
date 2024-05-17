package websocket

import (
	"context"
	"monitoring-system/internal/domain/camera"
	"monitoring-system/pkg/logger"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketServer struct {
	logger logger.Logger
}

func NewWebSocketServer(logger logger.Logger) *WebSocketServer {
	return &WebSocketServer{
		logger: logger,
	}
}

func (wss *WebSocketServer) streamVideo(ctx context.Context, cam camera.Camera, conn *websocket.Conn) {
	defer conn.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			img, err := cam.Capture()
			if err != nil {
				wss.logger.Error("Error capturing image: %v", err)
				return
			}

			err = conn.WriteMessage(websocket.BinaryMessage, img)
			if err != nil {
				wss.logger.Error("Error writing message: %v", err)
				return
			}
		}
	}
}

func (wss *WebSocketServer) VideoHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, cam camera.Camera) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		wss.logger.Error("Error upgrading to websocket: %v", err)
		return
	}
	go wss.streamVideo(ctx, cam, conn)
}
