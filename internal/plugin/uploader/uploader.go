package uploader

import "io"

type Uploader interface {
	PutObject(bucket string, key string, body io.Reader) error
	ObjectExists(bucket, key string) (bool, error)
	GetObject(bucket, key string) (io.ReadCloser, error)
	ListObjects(bucket, prefix string) ([]string, error)
	DeleteObject(bucket, key string) error
}
