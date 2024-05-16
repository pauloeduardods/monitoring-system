package camera

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"monitoring-system/pkg/logger"
	"time"

	"gocv.io/x/gocv"
)

type Webcam struct {
	deviceID int
	webcam   *gocv.VideoCapture
	logger   *logger.Logger
}

func NewWebcam(deviceID int, logger *logger.Logger) Camera {
	return &Webcam{deviceID: deviceID, logger: logger}
}

func (w *Webcam) Start() error {
	webcam, err := gocv.OpenVideoCapture(w.deviceID)
	if err != nil {
		return err
	}
	w.webcam = webcam
	return nil
}

func (w *Webcam) Stop() error {
	return w.webcam.Close()
}

func (w *Webcam) GetFPS() (float64, error) {
	fps := w.webcam.Get(gocv.VideoCaptureFPS)
	if fps == 0 {
		return 0, fmt.Errorf("unable to get FPS for device: %d", w.deviceID)
	}
	return fps, nil
}

func (w *Webcam) SetFPS(fps float64) error {
	w.webcam.Set(gocv.VideoCaptureFPS, fps)
	return nil
}

func (w *Webcam) Capture() (*gocv.Mat, error) {
	img := gocv.NewMat()
	if ok := w.webcam.Read(&img); !ok {
		return &img, fmt.Errorf("cannot read from device %d", w.deviceID)
	}
	if img.Empty() {
		return &img, fmt.Errorf("no image captured")
	}
	return &img, nil
}

func (c *Webcam) GetCapabilities() (CameraCapabilities, error) {
	width := c.webcam.Get(gocv.VideoCaptureFrameWidth)
	height := c.webcam.Get(gocv.VideoCaptureFrameHeight)
	if width == 0 || height == 0 {
		return CameraCapabilities{}, fmt.Errorf("unable to get dimensions for device: %d", c.deviceID)
	}
	fps := c.webcam.Get(gocv.VideoCaptureFPS)
	if fps == 0 {
		return CameraCapabilities{}, fmt.Errorf("unable to get FPS for device: %d", c.deviceID)
	}
	return CameraCapabilities{
		DeviceID: c.deviceID,
		Width:    int(width),
		Height:   int(height),
		FPS:      fps,
	}, nil
}

func (c *Webcam) Close() {
	c.webcam.Close()
}

func (w *Webcam) RecordVideo(ctx context.Context, filename string) error {
	capabilities, err := w.GetCapabilities()
	if err != nil {
		return err
	}

	writer, err := gocv.VideoWriterFile(filename, "MJPG", capabilities.FPS, capabilities.Width, capabilities.Height, true)
	if err != nil {
		return err
	}
	defer writer.Close()

	font := gocv.FontHersheyPlain
	scale := 1.5
	color := color.RGBA{R: 255, G: 255, B: 255, A: 0}
	thickness := 2
	position := image.Point{X: 10, Y: capabilities.Width - 10}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			img, err := w.Capture()
			if err != nil {
				w.logger.Warning("Error capturing image %v", err)
				continue
			}
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			gocv.PutText(img, timestamp, position, font, scale, color, thickness)

			writer.Write(*img)
		}
	}
}
