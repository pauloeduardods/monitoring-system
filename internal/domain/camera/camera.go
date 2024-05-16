package camera

type Camera interface {
	Start() error
	Stop() error
	GetDimensions() (int, int, error)
	RecordVideo(filename string, condition func() bool) error
}
