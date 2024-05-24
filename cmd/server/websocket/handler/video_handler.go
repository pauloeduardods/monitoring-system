package handler

import (
	"context"
	"monitoring-system/internal/domain/camera"
	"monitoring-system/pkg/logger"
	"net/http"

	"github.com/gorilla/websocket"
)

var WsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type VideoHandler interface {
	VideoHandler(w http.ResponseWriter, r *http.Request)
}

type videoHandler struct {
	camera camera.Camera
	ctx    context.Context
	logger logger.Logger
}

func NewVideoHandler(ctx context.Context, cam camera.Camera, logger logger.Logger) VideoHandler {
	return &videoHandler{
		camera: cam,
		ctx:    ctx,
		logger: logger,
	}
}

func (wss *videoHandler) streamVideo(ctx context.Context, cam camera.Camera, conn *websocket.Conn) {
	defer conn.Close()

	for {
		select {
		case <-ctx.Done():
		case <-cam.Done():
			return
		default:
			img, err := cam.Capture()
			if err != nil {
				wss.logger.Error("Error capturing image: %v", err)
				continue
			}
			err = conn.WriteMessage(websocket.BinaryMessage, img)
			if err != nil {
				wss.logger.Error("Error writing message: %v", err)
				return
			}
		}
	}
}

func (vh *videoHandler) VideoHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := WsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		vh.logger.Error("Error upgrading to websocket: %v", err)
		return
	}
	go vh.streamVideo(vh.ctx, vh.camera, conn)
}
