package camera

import (
	"context"
)

type Status string

const (
	Connected    Status = "connected"
	Disconnected Status = "disconnected"
	// Running      Status = "running"
	// Removed Status = "removed"
)

type Camera interface {
	Start() error
	Close() error
	RecordVideo(ctx context.Context, filename string) error
	Capture() ([]byte, error)
	StatusChan() <-chan Status
	GetDetails() CameraDetails
}

type CameraDetails struct {
	ID     int
	Name   string //TODO: Save name in the database
	Status Status
	Infos  Infos
}

type Infos struct {
	DeviceID int
	Width    int
	Height   int
	FPS      float64
}
