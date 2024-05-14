package camera

import (
	"time"

	"gocv.io/x/gocv"
)

func RecordVideo(camera Camera, filename string, duration time.Duration) error {
	writer, err := gocv.VideoWriterFile(filename, "MJPG", 20, 640, 480, true)
	if err != nil {
		return err
	}
	defer writer.Close()

	endTime := time.Now().Add(duration)
	for time.Now().Before(endTime) {
		img, err := camera.Capture()
		if err != nil {
			return err
		}
		writer.Write(img)
	}

	return nil
}
