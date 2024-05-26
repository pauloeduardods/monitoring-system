package camera

import (
	"context"
)

type Camera interface {
	Start() error
	Close() error
	RecordVideo(ctx context.Context, filename string) error
	Capture() ([]byte, error)
	Done() <-chan struct{}
	GetDetails() CameraDetails
}

type CameraDetails struct {
	ID    int
	Name  string //TODO: Save name in the database
	Infos Infos
}

type Infos struct {
	DeviceID int
	Width    int
	Height   int
	FPS      float64
}
