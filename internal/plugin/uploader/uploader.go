package uploader

import (
	"io"
	"time"
)

type Uploader interface {
	PutObject(bucket string, key string, body io.Reader) error
	ObjectExists(bucket, key string) (bool, error)
	GetObject(bucket, key string) (io.ReadCloser, error)
	ListObjects(bucket, prefix string) ([]string, error)
	DeleteObject(bucket, key string) error
	ListCommonPrefixes(bucket, prefix, delimiter string) ([]string, error)
	CreateSignedURL(bucketName, key string, ttl time.Duration) (string, error)
}
