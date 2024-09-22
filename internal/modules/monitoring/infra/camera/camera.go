package camera

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"monitoring-system/internal/modules/monitoring/domain/camera"
	"monitoring-system/pkg/logger"
	"time"

	"gocv.io/x/gocv"
)

type Camera struct {
	deviceID   int
	webcam     *gocv.VideoCapture
	logger     logger.Logger
	outputChan chan gocv.Mat
	details    *camera.CameraDetails
	ctx        context.Context
	cancel     context.CancelFunc
	done       chan struct{}
}

func NewCameraService(ctx context.Context, deviceID int, logger logger.Logger) camera.CameraService {
	cameraDetails := &camera.CameraDetails{
		ID:    deviceID,
		Name:  fmt.Sprintf("Camera %d", deviceID),
		Infos: camera.Infos{},
	}
	ctx, cancel := context.WithCancel(ctx)
	return &Camera{
		deviceID:   deviceID,
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
		outputChan: make(chan gocv.Mat),
		details:    cameraDetails,
		done:       make(chan struct{}),
	}
}

func (w *Camera) getInfos() (camera.Infos, error) {
	width := w.webcam.Get(gocv.VideoCaptureFrameWidth)
	height := w.webcam.Get(gocv.VideoCaptureFrameHeight)
	if width == 0 || height == 0 {
		return camera.Infos{}, fmt.Errorf("unable to get dimensions for device: %d", w.deviceID)
	}

	fps := w.webcam.Get(gocv.VideoCaptureFPS)
	if fps == 0 {
		return camera.Infos{}, fmt.Errorf("unable to get FPS for device: %d", w.deviceID)
	}

	return camera.Infos{
		DeviceID: w.deviceID,
		Width:    int(width),
		Height:   int(height),
		FPS:      fps,
	}, nil
}

func (w *Camera) Start() error {
	w.logger.Info("Starting webcam", w.deviceID)
	webcam, err := gocv.OpenVideoCapture(w.deviceID)
	if err != nil {
		return err
	}
	w.webcam = webcam
	infos, err := w.getInfos()
	if err != nil {
		return err
	}
	w.details.Infos = infos

	go w.capture()

	w.logger.Info("Camera started", w.deviceID)

	return nil
}

func (w *Camera) Close() error {
	defer close(w.outputChan)
	w.logger.Warning("Closing webcam", w.deviceID)
	w.cancel()
	close(w.done)
	return w.webcam.Close()
}

func (w *Camera) capture() {
	defer w.Close()

	maxRetries := 5
	retries := 0

	for {
		select {
		case <-w.done:
		case <-w.ctx.Done():
			w.logger.Warning("Camera capture stopped")
			return
		default:
			img := gocv.NewMat()

			ok := w.webcam.Read(&img)
			if !ok || img.Empty() {
				retries++
				if retries >= maxRetries {
					w.logger.Warning("Unable to read from device %d\n", w.deviceID)
					return
				}
				time.Sleep(1 * time.Second)

				img = gocv.NewMatWithSize(w.details.Infos.Height, w.details.Infos.Width, gocv.MatTypeCV8UC3)
				img.SetTo(gocv.NewScalar(0, 0, 0, 0))
			} else {
				retries = 0
			}

			font := gocv.FontHersheyPlain
			scale := 1.5
			color := color.RGBA{R: 255, G: 255, B: 255, A: 0}
			thickness := 2
			position := image.Point{X: 10, Y: w.details.Infos.Height - 10}

			timestamp := time.Now().Format("2006-01-02 15:04:05")
			gocv.PutText(&img, timestamp, position, font, scale, color, thickness)

			// TODO: Check if need to check for done or ctx.Done
			select {
			case <-w.done:
			case <-w.ctx.Done():
				return
			case w.outputChan <- img:
			}
			// w.outputChan <- img
		}
	}
}

func (w *Camera) Capture() ([]byte, error) {
	select {
	case <-w.done:
	case <-w.ctx.Done():
		w.logger.Info("Camera capture stopped", w.deviceID)
		return nil, nil
	case img := <-w.outputChan:
		buf, err := gocv.IMEncode(gocv.JPEGFileExt, img)
		if err != nil {
			w.logger.Error("Error encoding image: %v\n", err)
			return nil, err
		}
		defer img.Close()
		defer buf.Close()

		return buf.GetBytes(), nil
	}
	return nil, nil
}

func (w *Camera) RecordVideo(ctx context.Context, filename string) error {
	writer, err := gocv.VideoWriterFile(filename, "MJPG", w.details.Infos.FPS, w.details.Infos.Width, w.details.Infos.Height, true)
	if err != nil {
		return err
	}
	defer writer.Close()

	for {
		select {
		case <-w.done:
		case <-ctx.Done():
		case <-w.ctx.Done():
			return nil
		case img := <-w.outputChan:
			writer.Write(img)
		}
	}
}

func (w *Camera) GetDetails() camera.CameraDetails {
	return *w.details
}

func (w *Camera) Done() <-chan struct{} {
	return w.done
}
