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
	deviceID           int
	webcam             *gocv.VideoCapture
	logger             logger.Logger
	outputChan         chan gocv.Mat
	cameraCapabilities CameraCapabilities
	ctx                context.Context
}

func NewWebcam(ctx context.Context, deviceID int, logger logger.Logger) Camera {
	return &Webcam{deviceID: deviceID, logger: logger, ctx: ctx, outputChan: make(chan gocv.Mat)}
}

func (w *Webcam) getCameraCapabilities() (CameraCapabilities, error) {
	webcam, err := gocv.OpenVideoCapture(w.deviceID)
	if err != nil {
		return CameraCapabilities{}, err
	}
	defer webcam.Close()

	width := webcam.Get(gocv.VideoCaptureFrameWidth)
	height := webcam.Get(gocv.VideoCaptureFrameHeight)
	if width == 0 || height == 0 {
		return CameraCapabilities{}, fmt.Errorf("unable to get dimensions for device: %d", w.deviceID)
	}

	fps := webcam.Get(gocv.VideoCaptureFPS)
	if fps == 0 {
		return CameraCapabilities{}, fmt.Errorf("unable to get FPS for device: %d", w.deviceID)
	}

	return CameraCapabilities{
		DeviceID: w.deviceID,
		Width:    int(width),
		Height:   int(height),
		FPS:      fps,
	}, nil
}

func (w *Webcam) Check() (CameraCapabilities, error) {
	webcam, err := gocv.OpenVideoCapture(w.deviceID)
	if err != nil {
		return CameraCapabilities{}, err
	}
	defer webcam.Close()
	return w.getCameraCapabilities()
}

func (w *Webcam) Start() error {
	webcam, err := gocv.OpenVideoCapture(w.deviceID)
	if err != nil {
		return err
	}
	w.webcam = webcam

	capabilities, err := w.getCameraCapabilities()
	if err != nil {
		return err
	}
	w.cameraCapabilities = capabilities

	go w.capture()
	return nil
}

func (w *Webcam) Stop() error {
	err := w.webcam.Close()
	return err
}

func (w *Webcam) capture() {
	defer w.Stop()
	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info("Webcam capture stopped")
			return
		default:
			img := gocv.NewMat()

			ok := w.webcam.Read(&img)
			if !ok || img.Empty() {
				w.logger.Warning("Cannot read from device %d\n", w.deviceID)

				img = gocv.NewMatWithSize(w.cameraCapabilities.Height, w.cameraCapabilities.Width, gocv.MatTypeCV8UC3)
				img.SetTo(gocv.NewScalar(0, 0, 0, 0))
			}
			// if img.Empty() {
			// 	w.logger.Warning("Empty image from device %d\n", w.deviceID)

			// 	img = gocv.NewMatWithSize(w.cameraCapabilities.Height, w.cameraCapabilities.Width, gocv.MatTypeCV8UC3)
			// 	img.SetTo(gocv.NewScalar(0, 0, 0, 0))
			// }

			font := gocv.FontHersheyPlain
			scale := 1.5
			color := color.RGBA{R: 255, G: 255, B: 255, A: 0}
			thickness := 2
			position := image.Point{X: 10, Y: w.cameraCapabilities.Height - 10}

			timestamp := time.Now().Format("2006-01-02 15:04:05")
			gocv.PutText(&img, timestamp, position, font, scale, color, thickness)

			w.outputChan <- img
		}
	}
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

func (w *Webcam) Capture() ([]byte, error) {
	select {
	case <-w.ctx.Done():
		w.logger.Info("Webcam capture stopped")
		return nil, nil
	case img := <-w.outputChan:
		buf, err := gocv.IMEncode(gocv.JPEGFileExt, img)
		if err != nil {
			w.logger.Error("Error encoding image: %v\n", err)
			return nil, err
		}

		return buf.GetBytes(), nil
	}
}

func (c *Webcam) GetCapabilities() CameraCapabilities {
	return c.cameraCapabilities
}

func (c *Webcam) Close() {
	c.webcam.Close()
}

func (w *Webcam) RecordVideo(ctx context.Context, filename string) error {
	writer, err := gocv.VideoWriterFile(filename, "MJPG", w.cameraCapabilities.FPS, w.cameraCapabilities.Width, w.cameraCapabilities.Height, true)
	if err != nil {
		return err
	}
	defer writer.Close()

	for {
		select {
		case <-ctx.Done():
		case <-w.ctx.Done():
			return nil
		case img := <-w.outputChan:
			writer.Write(img)
		}
	}
}
