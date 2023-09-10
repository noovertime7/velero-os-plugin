package uploader

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"log"
)

// MinioUploader 实现了 Uploader 接口
type MinioUploader struct {
	client *minio.Client
}

// NewMinioUploader 创建一个 MinioUploader 实例
func NewMinioUploader(endpoint, accessKey, secretKey string, useSSL bool) (*MinioUploader, error) {
	// 创建 Minio 客户端
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &MinioUploader{client: minioClient}, nil
}

// PutObject 将数据上传到指定的桶和键中
func (u *MinioUploader) PutObject(bucket, key string, body io.Reader) error {
	_, err := u.client.PutObject(context.Background(), bucket, key, body, -1, minio.PutObjectOptions{})
	return err
}

// ObjectExists 检查指定的桶和键是否存在对象
func (u *MinioUploader) ObjectExists(bucket, key string) (bool, error) {
	_, err := u.client.StatObject(context.Background(), bucket, key, minio.StatObjectOptions{})
	if err != nil {
		if minioErr, ok := err.(minio.ErrorResponse); ok && minioErr.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetObject 获取指定桶和键的对象内容
func (u *MinioUploader) GetObject(bucket, key string) (io.ReadCloser, error) {
	obj, err := u.client.GetObject(context.Background(), bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// ListObjects 列出指定桶和前缀下的所有对象键
func (u *MinioUploader) ListObjects(bucket, prefix string) ([]string, error) {
	var objects []string

	for object := range u.client.ListObjects(context.Background(), bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
		if object.Err != nil {
			log.Println(object.Err)
			continue
		}

		objects = append(objects, object.Key)
	}

	return objects, nil
}
