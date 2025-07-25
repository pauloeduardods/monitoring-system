package camera

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"monitoring-system/src/config"
	"monitoring-system/src/internal/modules/monitoring/domain/camera"
	"monitoring-system/src/pkg/logger"
	"sync"
	"time"

	"gocv.io/x/gocv"
)

type Camera struct {
	id         string
	deviceID   interface{}
	webcam     *gocv.VideoCapture
	logger     logger.Logger
	outputChan chan gocv.Mat
	details    *camera.CameraDetails
	ctx        context.Context
	cancel     context.CancelFunc
	done       chan struct{}
	config     *config.CameraConfig
	closeOnce  sync.Once
}

func NewCameraService(ctx context.Context, id string, deviceID interface{}, logger logger.Logger, config *config.CameraConfig) camera.CameraService {
	if deviceID == nil {
		return nil
	}

	switch v := deviceID.(type) {
	case int:
		if v < 0 {
			logger.Error("deviceID is not a valid int")
			return nil
		}
	case string:
		if v == "" {
			logger.Error("deviceID is not a valid string")
			return nil
		}
	default:
		logger.Error("deviceID is not of type int or string")
		return nil
	}

	cameraDetails := &camera.CameraDetails{
		ID:    id,
		Name:  fmt.Sprintf("Camera %d", deviceID),
		Infos: camera.Infos{},
	}
	ctx, cancel := context.WithCancel(ctx)
	return &Camera{
		id:         id,
		deviceID:   deviceID,
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
		outputChan: make(chan gocv.Mat),
		details:    cameraDetails,
		done:       make(chan struct{}),
		config:     config,
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
	// w.logger.Info("Starting webcam", w.deviceID)
	webcam, err := gocv.OpenVideoCapture(w.deviceID)
	if err != nil {
		return err
	}

	if !webcam.IsOpened() {
		return fmt.Errorf("error starting webcam device %d", w.deviceID)
	}

	w.webcam = webcam
	infos, err := w.getInfos()
	if err != nil {
		return err
	}

	if infos.FPS <= 0 {
		return fmt.Errorf("error starting webcam device %d fps: %f", w.deviceID, infos.FPS)
	}

	webcam.Set(gocv.VideoCaptureFrameWidth, float64(w.config.Width))
	webcam.Set(gocv.VideoCaptureFrameHeight, float64(w.config.Height))
	webcam.Set(gocv.VideoCaptureFPS, float64(w.config.FPS))
	webcam.Set(gocv.VideoCaptureFOURCC, float64(webcam.ToCodec(w.config.Codec)))

	w.details.Infos = infos

	go w.capture()

	w.logger.Info("Camera started", w.deviceID)

	return nil
}

func (w *Camera) Close() error {
	w.logger.Warning("Closing webcam", w.deviceID)
	w.cancel()
	w.closeOnce.Do(func() {
		close(w.done)
	})
	if w.webcam != nil {
		return w.webcam.Close()
	}
	return nil
}

func (w *Camera) capture() {
	defer w.Close()

	maxRetries := 5
	retries := 0

	for {
		select {
		case <-w.done:
			w.logger.Info("Capture loop done for device", w.deviceID)
			return
		case <-w.ctx.Done():
			w.logger.Warning("Context canceled, stopping capture")
			return
		default:
			img := gocv.NewMat()

			if ok := w.webcam.Read(&img); !ok || img.Empty() {
				img.Close()
				retries++
				if retries >= maxRetries {
					w.logger.Warning("Unable to read from device %d after %d retries\n", w.deviceID, retries)
					return
				}
				// w.logger.Warning("Retrying capture for device %d", w.deviceID)
				time.Sleep(1 * time.Second)
				continue
			}
			retries = 0

			font := gocv.FontHersheyPlain
			scale := 1.5
			color := color.RGBA{R: 255, G: 255, B: 255, A: 0}
			thickness := 2
			position := image.Point{X: 10, Y: w.details.Infos.Height - 10}

			timestamp := time.Now().Format("2006-01-02 15:04:05")
			gocv.PutText(&img, timestamp, position, font, scale, color, thickness)

			select {
			case <-w.done:
				w.logger.Warning("Capture stopped by done signal")
				return
			case <-w.ctx.Done():
				w.logger.Warning("Capture stopped by context cancellation")
				return
			case w.outputChan <- img:
				// Image sent successfully
			}
		}
	}
}

func (w *Camera) Capture() ([]byte, error) {
	select {
	case <-w.done:
		w.logger.Info("Camera done signal received, stopping capture", w.deviceID)
		return nil, nil
	case <-w.ctx.Done():
		w.logger.Info("Context done, stopping capture", w.deviceID)
		return nil, nil
	case img := <-w.outputChan:
		defer img.Close()
		image, err := img.ToImage()
		if err != nil {
			w.logger.Error("Error to get image.Image from gocv.Mat", err)
			return nil, err
		}

		buffer := new(bytes.Buffer)
		if err := jpeg.Encode(buffer, image, &jpeg.Options{Quality: 75}); err != nil {
			w.logger.Error("Error encoding image", err)
			return nil, err
		}

		return buffer.Bytes(), nil
	}
}

func (w *Camera) RecordVideo(ctx context.Context, filename string, motionOnly bool) error {
	writer, err := gocv.VideoWriterFile(filename, w.config.Codec, float64(w.config.FPS), w.config.Width, w.config.Height, w.config.MotionDetection)
	if err != nil {
		return err
	}
	defer writer.Close()

	imgDelta := gocv.NewMat()
	defer imgDelta.Close()

	imgThresh := gocv.NewMat()
	defer imgThresh.Close()

	mog2 := gocv.NewBackgroundSubtractorMOG2()
	defer mog2.Close()

	for {
		select {
		case <-w.done:
			w.logger.Info("Recording done for device", w.deviceID)
			return nil
		case <-ctx.Done():
			w.logger.Warning("Recording stopped by context cancellation")
			return nil
		case img := <-w.outputChan:
			if motionOnly {
				mog2.Apply(img, &imgDelta)

				gocv.Threshold(imgDelta, &imgThresh, 25, 255, gocv.ThresholdBinary)

				kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
				gocv.Dilate(imgThresh, &imgThresh, kernel)
				kernel.Close()

				contours := gocv.FindContours(imgThresh, gocv.RetrievalExternal, gocv.ChainApproxSimple)

				for i := 0; i < contours.Size(); i++ {
					area := gocv.ContourArea(contours.At(i))
					if area < float64(w.config.MinArea) {
						continue
					}
					err := writer.Write(img)
					if err != nil {
						w.logger.Error("Error while writing frame")
					}
				}
			} else {
				err := writer.Write(img)
				if err != nil {
					w.logger.Error("Error while writing frame")
				}
			}
		}
	}
}

func (w *Camera) GetDetails() camera.CameraDetails {
	return *w.details
}

func (w *Camera) Done() <-chan struct{} {
	return w.done
}
