package camera

import (
	"context"

	"gocv.io/x/gocv"
)

type Camera interface {
	Start() error
	Stop() error
	GetDimensions() (int, int, error)
	RecordVideo(ctx context.Context, filename string) error
	Capture() (gocv.Mat, error)
}
