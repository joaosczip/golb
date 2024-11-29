package config

import "os"

type FileReader interface {
	Read(path string) ([]byte, error)
}

type OSFileReader struct {}

func (o *OSFileReader) Read(path string) ([]byte, error) {
	return os.ReadFile(path)
}