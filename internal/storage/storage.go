package storage

import "errors"

type Storage interface {
	Save(key string, data []byte) error
	Get(key string) ([]byte, error)
}

var ErrFileNotFound = errors.New("file not found")
