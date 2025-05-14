package camera

import (
	"context"
)

type CameraService interface {
	Start() error
	Close() error
	RecordVideo(ctx context.Context, filename string, motionOnly bool) error
	Capture() ([]byte, error)
	Done() <-chan struct{}
	GetDetails() CameraDetails
}

type CameraDetails struct {
	ID    string
	Name  string
	Infos Infos
}

type Infos struct {
	DeviceID interface{}
	Width    int
	Height   int
	FPS      float64
}
