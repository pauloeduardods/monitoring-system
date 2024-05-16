package camera

import (
	"context"
)

type Camera interface {
	Start() error
	Stop() error
	GetCapabilities() CameraCapabilities
	SetFPS(fps float64) error
	RecordVideo(ctx context.Context, filename string) error
	Capture() ([]byte, error)
}

type CameraCapabilities struct {
	DeviceID int
	Width    int
	Height   int
	FPS      float64
}
