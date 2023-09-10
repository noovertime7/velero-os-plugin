package uploader

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io"
)

// OSSUploader 实现了 Uploader 接口
type OSSUploader struct {
	client *oss.Client
	bucket *oss.Bucket
}

// NewOSSUploader 创建一个 OSSUploader 实例
func NewOSSUploader(endpoint, accessKey, secretKey, bucketName string) (*OSSUploader, error) {
	// 创建 OSS 客户端
	client, err := oss.New(endpoint, accessKey, secretKey)
	if err != nil {
		return nil, err
	}

	// 获取存储空间
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}

	return &OSSUploader{
		client: client,
		bucket: bucket,
	}, nil
}

// PutObject 将数据上传到指定的桶和键中
func (u *OSSUploader) PutObject(bucket, key string, body io.Reader) error {
	return u.bucket.PutObject(key, body)
}

// ObjectExists 检查指定的桶和键是否存在对象
func (u *OSSUploader) ObjectExists(bucket, key string) (bool, error) {
	return u.bucket.IsObjectExist(key)
}

// GetObject 获取指定桶和键的对象内容
func (u *OSSUploader) GetObject(bucket, key string) (io.ReadCloser, error) {
	return u.bucket.GetObject(key)
}

// ListObjects 列出指定桶和前缀下的所有对象键
func (u *OSSUploader) ListObjects(bucket, prefix string) ([]string, error) {
	result, err := u.bucket.ListObjects(oss.Prefix(prefix))
	if err != nil {
		return nil, err
	}

	var objects []string
	for _, object := range result.Objects {
		objects = append(objects, object.Key)
	}

	return objects, nil
}

// DeleteObject 删除指定桶和键的对象
func (u *OSSUploader) DeleteObject(bucket, key string) error {
	return u.bucket.DeleteObject(key)
}
