package camera

import (
	"fmt"

	"gocv.io/x/gocv"
)

type Camera interface {
	Start() error
	Stop() error
	Capture() (gocv.Mat, error)
}

type Webcam struct {
	deviceID int
	webcam   *gocv.VideoCapture
}

func NewWebcam(deviceID int) *Webcam {
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
