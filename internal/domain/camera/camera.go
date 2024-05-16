package camera

import (
	"context"

	"gocv.io/x/gocv"
)

type Camera interface {
	Start() error
	Stop() error
	GetCapabilities() (CameraCapabilities, error)
	RecordVideo(ctx context.Context, filename string) error
	Capture() (*gocv.Mat, error)
	SetFPS(fps float64) error
}

type CameraCapabilities struct {
	DeviceID int
	Width    int
	Height   int
	FPS      float64
}
