package handler

import (
	"context"
	"monitoring-system/src/internal/modules/monitoring/domain/camera"
	"monitoring-system/src/pkg/logger"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	FPS_STREAM_LIMIT = 10
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
	camera camera.CameraService
	ctx    context.Context
	logger logger.Logger
}

func NewVideoHandler(ctx context.Context, cam camera.CameraService, logger logger.Logger) VideoHandler {
	return &videoHandler{
		camera: cam,
		ctx:    ctx,
		logger: logger,
	}
}

func (wss *videoHandler) streamVideo(ctx context.Context, cam camera.CameraService, conn *websocket.Conn) {
	defer func() {
		if r := recover(); r != nil {
			wss.logger.Error("Recovered from panic in streamVideo: %v", r)
			conn.Close()
		}
	}()

	defer conn.Close()

	frameInterval := time.Second / time.Duration(FPS_STREAM_LIMIT)

	ticker := time.NewTicker(frameInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cam.Done():
			return
		case <-ticker.C:
			img, err := cam.Capture()
			if err != nil {
				wss.logger.Error("Error capturing image from camera %d: %v", wss.camera.GetDetails().ID, err)
				continue
			}
			if len(img) == 0 {
				wss.logger.Error("Empty image captured from camera %d", wss.camera.GetDetails().ID)
				continue
			}
			err = conn.WriteMessage(websocket.BinaryMessage, img)
			if err != nil {
				wss.logger.Error("Error sending image through WebSocket for camera %d: %v", wss.camera.GetDetails().ID, err)
				conn.Close()
				return
			}
		}
	}
}

func (vh *videoHandler) VideoHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := WsUpgrader.Upgrade(w, r, nil)
	if err != nil || conn == nil {
		vh.logger.Error("Error upgrading to websocket: %v", err, r)
		return
	}
	go vh.streamVideo(vh.ctx, vh.camera, conn)
}
