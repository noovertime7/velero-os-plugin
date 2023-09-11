package uploader

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/sirupsen/logrus"
	"io"
	"time"
)

// OSSUploader 实现了 Uploader 接口
type OSSUploader struct {
	client *oss.Client
	log    logrus.FieldLogger
}

// NewOSSUploader 创建一个 OSSUploader 实例
func NewOSSUploader(endpoint, accessKey, secretKey, region string, s3ForcePathStyle bool, log logrus.FieldLogger) (Uploader, error) {
	// 创建 OSS 客户端
	client, err := oss.New(endpoint, accessKey, secretKey, oss.Region(region), oss.ForcePathStyle(s3ForcePathStyle))
	if err != nil {
		return nil, err
	}

	log.Info("build oss uploader success")

	return &OSSUploader{
		client: client,
		log:    log,
	}, nil
}

// PutObject 将数据上传到指定的桶和键中
func (o *OSSUploader) PutObject(bucketName, key string, body io.Reader) error {
	// 获取存储空间
	bucket, err := o.client.Bucket(bucketName)
	if err != nil {
		return err
	}
	return bucket.PutObject(key, body)
}

// ObjectExists 检查指定的桶和键是否存在对象
func (o *OSSUploader) ObjectExists(bucketName, key string) (bool, error) {
	// 获取存储空间
	bucket, err := o.client.Bucket(bucketName)
	if err != nil {
		return false, err
	}
	return bucket.IsObjectExist(key)
}

// GetObject 获取指定桶和键的对象内容
func (o *OSSUploader) GetObject(bucketName, key string) (io.ReadCloser, error) {
	// 获取存储空间
	bucket, err := o.client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}
	return bucket.GetObject(key)
}

// ListObjects 列出指定桶和前缀下的所有对象键
func (o *OSSUploader) ListObjects(bucketName, prefix string) ([]string, error) {
	bucket, err := o.client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}
	result, err := bucket.ListObjects(oss.Prefix(prefix))
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
func (o *OSSUploader) DeleteObject(bucketName, key string) error {
	bucket, err := o.client.Bucket(bucketName)
	if err != nil {
		return err
	}
	return bucket.DeleteObject(key)
}

func (o *OSSUploader) ListCommonPrefixes(bucketName, prefix, delimiter string) ([]string, error) {
	bucket, err := o.client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}

	prefixes := make([]string, 0)

	marker := ""
	for {
		options := []oss.Option{
			oss.Prefix(prefix),
			oss.Marker(marker),
			oss.Delimiter(delimiter),
		}

		lsRes, err := bucket.ListObjects(options...)
		if err != nil {
			return nil, err
		}

		for _, commonPrefix := range lsRes.CommonPrefixes {
			prefixes = append(prefixes, commonPrefix)
		}

		if !lsRes.IsTruncated {
			break
		}
		marker = lsRes.NextMarker
	}

	return prefixes, nil
}

func (o *OSSUploader) CreateSignedURL(bucketName, key string, ttl time.Duration) (string, error) {
	bucket, err := o.client.Bucket(bucketName)
	if err != nil {
		return "", err
	}

	// 生成过期时间
	expireTime := time.Now().Add(ttl)

	// 生成预签名 URL
	signedURL, err := bucket.SignURL(key, oss.HTTPGet, int64(expireTime.Unix()))
	if err != nil {
		return "", err
	}

	return signedURL, nil
}
