/*
Copyright 2017, 2019 the Velero contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plugin

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/pkg/errors"
	"github.com/vmware-tanzu/velero-plugin-example/internal/ini"
	"github.com/vmware-tanzu/velero-plugin-example/internal/plugin/uploader"
	veleroplugin "github.com/vmware-tanzu/velero/pkg/plugin/framework"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	s3TypeKey                = "s3Type"
	regionKey                = "region"
	s3URLKey                 = "s3Url"
	insecureSkipTLSVerifyKey = "insecureSkipTLSVerify"
	s3ForcePathStyleKey      = "s3ForcePathStyle"
	bucketKey                = "bucket"
	credentialsFileKey       = "credentialsFile"
	credentialProfileKey     = "profile"
)

type ObjectStore struct {
	log      logrus.FieldLogger
	uploader uploader.Uploader
}

func NewObjectStore(logger logrus.FieldLogger) *ObjectStore {
	return &ObjectStore{log: logger}
}

// Init initializes the plugin. After v0.10.0, this can be called multiple times.
func (f *ObjectStore) Init(config map[string]string) error {
	f.log.Infof("Init called")
	f.log.Debugf("config:[%v]", config)
	if err := veleroplugin.ValidateObjectStoreConfigKeys(config,
		regionKey,
		s3URLKey,
		s3TypeKey,
		s3ForcePathStyleKey,
		credentialsFileKey,
		credentialProfileKey,
		insecureSkipTLSVerifyKey,
	); err != nil {
		return err
	}

	var (
		region                   = config[regionKey]
		s3URL                    = config[s3URLKey]
		s3Type                   = config[s3TypeKey]
		insecureSkipTLSVerifyVal = config[insecureSkipTLSVerifyKey]
		s3ForcePathStyleVal      = config[s3ForcePathStyleKey]
		credentialProfile        = config[credentialProfileKey]
		credentialsFile          = config[credentialsFileKey]
		bucket                   = config[bucketKey]
		s3ForcePathStyle         bool
		insecureSkipTLSVerify    bool
		err                      error
	)
	f.log.Info("bucket", bucket)

	if insecureSkipTLSVerifyVal != "" {
		if insecureSkipTLSVerify, err = strconv.ParseBool(insecureSkipTLSVerifyVal); err != nil {
			return errors.Wrapf(err, "could not parse %s (expected bool)", insecureSkipTLSVerifyKey)
		}
	}

	if s3ForcePathStyleVal != "" {
		if s3ForcePathStyle, err = strconv.ParseBool(s3ForcePathStyleVal); err != nil {
			return errors.Wrapf(err, "could not parse %s (expected bool)", s3ForcePathStyleKey)
		}
	}

	access, secret, err := f.getAccessAndSecret(credentialsFile, credentialProfile)
	if err != nil {
		return err
	}

	switch s3Type {
	case "minio":
		f.uploader, err = uploader.NewMinioUploader(s3URL, access, secret, insecureSkipTLSVerify, region, f.log)
		if err != nil {
			return fmt.Errorf("init minio uploader error: %w", err)
		}
	case "oss":
		f.uploader, err = uploader.NewOSSUploader(s3URL, access, secret, region, s3ForcePathStyle, f.log)
		if err != nil {
			return fmt.Errorf("init oss uploader error: %w", err)
		}
	default:
		return fmt.Errorf("unsurport s3 Type")
	}

	f.log.Debugf("build os-plugin uploader success,uploader type: [%s]", s3Type)

	return nil
}

func (f *ObjectStore) getAccessAndSecret(credentialsFile, profile string) (string, string, error) {
	if len(profile) == 0 {
		profile = DefaultSharedConfigProfile
	}

	if credentialsFile != "" {
		if _, err := os.Stat(credentialsFile); err != nil {
			if os.IsNotExist(err) {
				return "", "", errors.Wrapf(err, "provided credentialsFile does not exist")
			}
			return "", "", errors.Wrapf(err, "could not get credentialsFile info")
		}

		//	从给出的配置文件路径中读取用户名密码
		f.log.Infof("从给出的配置中读取密钥: [%s] [%s]", credentialsFile, profile)
		access, secret, err := f.readCredentialsFile(credentialsFile, profile)
		if err != nil {
			return "", "", err
		}
		return access, secret, nil
	}
	f.log.Infof("从默认配置中读取密钥: [%s] [%s]", DefaultCredentialsFile, profile)
	return f.readCredentialsFile(DefaultCredentialsFile, profile)
}

func (f *ObjectStore) readCredentialsFile(CredentialsFile, profile string) (string, string, error) {
	f.log.Infof("read CredentialsFile from: %s", CredentialsFile)
	config, err := ini.OpenFile(CredentialsFile)
	if err != nil {
		return "", "", fmt.Errorf("read from %s error", DefaultCredentialsFile)
	}
	iniProfile, ok := config.GetSection(profile)
	if !ok {
		return "", "", awserr.New("SharedCredsLoad", "failed to get profile", nil)
	}

	id := iniProfile.String(accessKeyIDKey)
	if len(id) == 0 {
		return "", "", fmt.Errorf("get aws_access_key_id empty :%v", id)
	}

	secret := iniProfile.String(secretAccessKey)
	if len(secret) == 0 {
		return "", "", fmt.Errorf("get aws_secret_access_key empty :%v", id)
	}

	return id, secret, nil
}

func (f *ObjectStore) PutObject(bucket string, key string, body io.Reader) error {
	f.log.Infof("PutObject  [%s/%s]", bucket, key)
	return f.uploader.PutObject(bucket, key, body)
}

func (f *ObjectStore) ObjectExists(bucket, key string) (bool, error) {
	f.log.Infof("check object exists  [%s/%s]", bucket, key)
	return f.uploader.ObjectExists(bucket, key)
}

func (f *ObjectStore) GetObject(bucket, key string) (io.ReadCloser, error) {
	f.log.Infof("get object [%s/%s]", bucket, key)
	return f.uploader.GetObject(bucket, key)
}

func (f *ObjectStore) ListCommonPrefixes(bucket, prefix, delimiter string) ([]string, error) {
	f.log.Infof("ListCommonPrefixes object [%s/%s/%s]", bucket, prefix, delimiter)
	return f.uploader.ListCommonPrefixes(bucket, prefix, delimiter)
}

func (f *ObjectStore) ListObjects(bucket, prefix string) ([]string, error) {
	f.log.Infof("list object [%s/%s]", bucket, prefix)
	return f.uploader.ListObjects(bucket, prefix)
}

func (f *ObjectStore) DeleteObject(bucket, key string) error {
	f.log.Infof("delete object [%s/%s]", bucket, key)
	return f.uploader.DeleteObject(bucket, key)
}

func (f *ObjectStore) CreateSignedURL(bucket, key string, ttl time.Duration) (string, error) {
	log := f.log.WithFields(logrus.Fields{
		"bucket": bucket,
		"key":    key,
		"ttl":    ttl,
	})
	log.Infof("CreateSignedURL")
	return f.uploader.CreateSignedURL(bucket, key, ttl)
}
