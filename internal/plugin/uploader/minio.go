package uploader

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
	"io"
	"path/filepath"
	"strings"
	"time"
)

// MinioUploader 实现了 Uploader 接口
type MinioUploader struct {
	client *minio.Client
	logger logrus.FieldLogger
}

func (m *MinioUploader) DeleteObject(bucket, key string) error {
	logrus.Debugf("delete object [%s/%s]", bucket, key)
	return m.client.RemoveObject(context.Background(), bucket, key, minio.RemoveObjectOptions{})
}

// NewMinioUploader 创建一个 MinioUploader 实例
func NewMinioUploader(endpoint, accessKey, secretKey string, useSSL bool, region string, logger logrus.FieldLogger) (Uploader, error) {
	if strings.HasPrefix(endpoint, "http://") {
		endpoint = strings.TrimPrefix(endpoint, "http://")
	} else if strings.HasPrefix(endpoint, "https://") {
		endpoint = strings.TrimPrefix(endpoint, "https://")
	}
	// 创建 Minio 客户端
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		return nil, err
	}

	logger.Info("build minio uploader success")
	return &MinioUploader{client: minioClient, logger: logger}, nil
}

// PutObject 将数据上传到指定的桶和键中
func (m *MinioUploader) PutObject(bucket, key string, body io.Reader) error {
	m.logger.Debugf("upload object [%s/%s]", bucket, key)
	_, err := m.client.PutObject(context.Background(), bucket, key, body, -1, minio.PutObjectOptions{})
	return err
}

// ObjectExists 检查指定的桶和键是否存在对象
func (m *MinioUploader) ObjectExists(bucket, key string) (bool, error) {
	_, err := m.client.StatObject(context.Background(), bucket, key, minio.StatObjectOptions{})
	if err != nil {
		if minioErr, ok := err.(minio.ErrorResponse); ok && minioErr.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetObject 获取指定桶和键的对象内容
func (m *MinioUploader) GetObject(bucket, key string) (io.ReadCloser, error) {
	obj, err := m.client.GetObject(context.Background(), bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// ListObjects 列出指定桶和前缀下的所有对象键
func (m *MinioUploader) ListObjects(bucket, prefix string) ([]string, error) {
	var objects []string

	ListObjectsOptions := minio.ListObjectsOptions{}
	if prefix != "" {
		ListObjectsOptions = minio.ListObjectsOptions{Prefix: prefix, Recursive: true}
	}

	for object := range m.client.ListObjects(context.Background(), bucket, ListObjectsOptions) {
		if object.Err != nil {
			m.logger.Error(object.Err)
			continue
		}

		objects = append(objects, object.Key)
	}

	return objects, nil
}

func (m *MinioUploader) ListCommonPrefixes(bucket, prefix, delimiter string) ([]string, error) {
	ctx := context.Background()
	prefixes := make([]string, 0)

	ListObjectsOptions := minio.ListObjectsOptions{}
	if prefix != "" {
		ListObjectsOptions = minio.ListObjectsOptions{Prefix: prefix, Recursive: true}
	}
	// 列举对象
	objectCh := m.client.ListObjects(ctx, bucket, ListObjectsOptions)

	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}

		// 提取公共前缀
		objectKey := object.Key

		dir := filepath.Dir(objectKey[len(prefix):])
		commonPrefix := fmt.Sprintf(prefix + dir + delimiter)

		// 如果该前缀已经存在于 prefixes 切片中，则跳过
		if contains(prefixes, commonPrefix) {
			continue
		}

		// 将公共前缀添加到切片中
		prefixes = append(prefixes, commonPrefix)
	}

	return prefixes, nil
}

func (m *MinioUploader) CreateSignedURL(bucket, key string, ttl time.Duration) (string, error) {
	// 创建预签名 URL
	presignedURL, err := m.client.PresignedGetObject(context.Background(), bucket, key, ttl, nil)
	if err != nil {
		return "", err
	}

	return presignedURL.String(), nil
}

func contains(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}
