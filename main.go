package main

import (
	"fmt"
	"time"

	"gocv.io/x/gocv"
)

func captureVideoFromCamera(deviceID int, filename string, duration time.Duration) error {
	webcam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		return fmt.Errorf("error opening video capture device: %v", err)
	}
	defer webcam.Close()

	writer, err := gocv.VideoWriterFile(filename, "MJPG", 20, 640, 480, true)
	if err != nil {
		return fmt.Errorf("error opening video writer: %v", err)
	}
	defer writer.Close()

	img := gocv.NewMat()
	defer img.Close()

	endTime := time.Now().Add(duration)
	for time.Now().Before(endTime) {
		if ok := webcam.Read(&img); !ok {
			return fmt.Errorf("cannot read device %d", deviceID)
		}
		if img.Empty() {
			continue
		}

		writer.Write(img)
	}

	return nil
}

func main() {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("camera_video_%s.avi", timestamp)

	duration := 5 * time.Second

	err := captureVideoFromCamera(0, filename, duration)
	if err != nil {
		fmt.Printf("Error capturing video from camera: %v\n", err)
	} else {
		fmt.Printf("Video saved as %s\n", filename)
	}
}
