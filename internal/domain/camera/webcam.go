package camera

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"time"

	"gocv.io/x/gocv"
)

type Webcam struct {
	deviceID int
	webcam   *gocv.VideoCapture
}

func NewWebcam(deviceID int) Camera {
	return &Webcam{deviceID: deviceID}
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

func (w *Webcam) Capture() (gocv.Mat, error) {
	img := gocv.NewMat()
	if ok := w.webcam.Read(&img); !ok {
		return img, fmt.Errorf("cannot read from device %d", w.deviceID)
	}
	if img.Empty() {
		return img, fmt.Errorf("no image captured")
	}
	return img, nil
}

func (c *Webcam) GetDimensions() (int, int, error) {
	width := c.webcam.Get(gocv.VideoCaptureFrameWidth)
	height := c.webcam.Get(gocv.VideoCaptureFrameHeight)
	if width == 0 || height == 0 {
		return 0, 0, fmt.Errorf("unable to get dimensions for device: %d", c.deviceID)
	}
	return int(width), int(height), nil
}

func (c *Webcam) Close() {
	c.webcam.Close()
}

func (w *Webcam) RecordVideo(ctx context.Context, filename string) error {
	width, height, err := w.GetDimensions()
	if err != nil {
		return err
	}

	writer, err := gocv.VideoWriterFile(filename, "MJPG", 20, width, height, true)
	if err != nil {
		return err
	}
	defer writer.Close()

	font := gocv.FontHersheyPlain
	scale := 1.5
	color := color.RGBA{R: 255, G: 255, B: 255, A: 0}
	thickness := 2
	position := image.Point{X: 10, Y: height - 10}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			img, err := w.Capture()
			if err != nil {
				return err
			}
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			gocv.PutText(&img, timestamp, position, font, scale, color, thickness)

			writer.Write(img)
		}
	}
}
