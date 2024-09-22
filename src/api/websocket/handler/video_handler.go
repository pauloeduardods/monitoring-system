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
	defer conn.Close()

	frameInterval := time.Second / time.Duration(FPS_STREAM_LIMIT)

	lastFrameTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cam.Done():
			return
		default:

			elapsed := time.Since(lastFrameTime)
			if elapsed < frameInterval {

				time.Sleep(frameInterval - elapsed)
				continue
			}

			lastFrameTime = time.Now()

			img, err := cam.Capture()
			if err != nil || img == nil {
				wss.logger.Error("Error capturing image: %v", err)
				continue
			}

			if conn == nil {
				wss.logger.Error("WebSocket connection is nil")
				return
			}
			if len(img) == 0 {
				wss.logger.Error("Image is Empty")
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
	if err != nil || conn == nil {
		vh.logger.Error("Error upgrading to websocket: %v", err)
		return
	}
	go vh.streamVideo(vh.ctx, vh.camera, conn)
}
