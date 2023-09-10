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
	"errors"
	"github.com/vmware-tanzu/velero-plugin-example/internal/plugin/uploader"
	veleroplugin "github.com/vmware-tanzu/velero/pkg/plugin/framework"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	regionKey                    = "region"
	s3URLKey                     = "s3Url"
	publicURLKey                 = "publicUrl"
	kmsKeyIDKey                  = "kmsKeyId"
	customerKeyEncryptionFileKey = "customerKeyEncryptionFile"
	s3ForcePathStyleKey          = "s3ForcePathStyle"
	bucketKey                    = "bucket"
	signatureVersionKey          = "signatureVersion"
	credentialsFileKey           = "credentialsFile"
	credentialProfileKey         = "profile"
	serverSideEncryptionKey      = "serverSideEncryption"
	insecureSkipTLSVerifyKey     = "insecureSkipTLSVerify"
	caCertKey                    = "caCert"
	enableSharedConfigKey        = "enableSharedConfig"
)

type ObjectStore struct {
	log      logrus.FieldLogger
	uploader uploader.Uploader
}

func newObjectStore(logger logrus.FieldLogger) *ObjectStore {
	return &ObjectStore{log: logger}
}

// NewFileObjectStore instantiates a FileObjectStore.
func NewFileObjectStore(log logrus.FieldLogger) *FileObjectStore {
	return &FileObjectStore{log: log}
}

// Init initializes the plugin. After v0.10.0, this can be called multiple times.
func (f *ObjectStore) Init(config map[string]string) error {
	f.log.Infof("FileObjectStore.Init called")

	if err := veleroplugin.ValidateObjectStoreConfigKeys(config,
		regionKey,
		s3URLKey,
		publicURLKey,
		kmsKeyIDKey,
		customerKeyEncryptionFileKey,
		s3ForcePathStyleKey,
		signatureVersionKey,
		credentialsFileKey,
		credentialProfileKey,
		serverSideEncryptionKey,
		insecureSkipTLSVerifyKey,
		enableSharedConfigKey,
	); err != nil {
		return err
	}

	var (
		region                    = config[regionKey]
		s3URL                     = config[s3URLKey]
		publicURL                 = config[publicURLKey]
		kmsKeyID                  = config[kmsKeyIDKey]
		customerKeyEncryptionFile = config[customerKeyEncryptionFileKey]
		s3ForcePathStyleVal       = config[s3ForcePathStyleKey]
		signatureVersion          = config[signatureVersionKey]
		credentialProfile         = config[credentialProfileKey]
		credentialsFile           = config[credentialsFileKey]
		serverSideEncryption      = config[serverSideEncryptionKey]
		insecureSkipTLSVerifyVal  = config[insecureSkipTLSVerifyKey]
		enableSharedConfig        = config[enableSharedConfigKey]

		// note that bucket is automatically added to the config map
		// by the server from the ObjectStorageProviderConfig so
		// doesn't need to be explicitly set by the user within
		// config.
		bucket                = config[bucketKey]
		caCert                = config[caCertKey]
		s3ForcePathStyle      bool
		insecureSkipTLSVerify bool
		err                   error
	)

}

func (f *FileObjectStore) PutObject(bucket string, key string, body io.Reader) error {
	path := filepath.Join(getRoot(), bucket, key)

	log := f.log.WithFields(logrus.Fields{
		"bucket": bucket,
		"key":    key,
		"path":   path,
	})
	log.Infof("PutObject")

	dir := filepath.Dir(path)
	log.Infof("Creating dir %s", dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	log.Infof("Creating file")
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	log.Infof("Writing to file")
	_, err = io.Copy(file, body)

	log.Infof("Done")
	return err
}

func (f *FileObjectStore) ObjectExists(bucket, key string) (bool, error) {
	path := filepath.Join(getRoot(), bucket, key)

	log := f.log.WithFields(logrus.Fields{
		"bucket": bucket,
		"key":    key,
		"path":   path,
	})
	log.Infof("ObjectExists")

	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}

func (f *FileObjectStore) GetObject(bucket, key string) (io.ReadCloser, error) {
	path := filepath.Join(getRoot(), bucket, key)

	log := f.log.WithFields(logrus.Fields{
		"bucket": bucket,
		"key":    key,
		"path":   path,
	})
	log.Infof("GetObject")

	return os.Open(path)
}

func (f *FileObjectStore) ListCommonPrefixes(bucket, prefix, delimiter string) ([]string, error) {
	path := filepath.Join(getRoot(), bucket, prefix, delimiter)

	log := f.log.WithFields(logrus.Fields{
		"bucket":    bucket,
		"delimiter": delimiter,
		"path":      path,
		"prefix":    prefix,
	})
	log.Infof("ListCommonPrefixes")

	infos, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, info := range infos {
		if info.IsDir() {
			dirs = append(dirs, info.Name())
		}
	}

	return dirs, nil
}

func (f *FileObjectStore) ListObjects(bucket, prefix string) ([]string, error) {
	path := filepath.Join(getRoot(), bucket, prefix)

	log := f.log.WithFields(logrus.Fields{
		"bucket": bucket,
		"prefix": prefix,
		"path":   path,
	})
	log.Infof("ListObjects")

	infos, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var objects []string
	for _, info := range infos {
		objects = append(objects, filepath.Join(prefix, info.Name()))
	}

	return objects, nil
}

func (f *FileObjectStore) DeleteObject(bucket, key string) error {
	path := filepath.Join(getRoot(), bucket, key)

	log := f.log.WithFields(logrus.Fields{
		"bucket": bucket,
		"key":    key,
		"path":   path,
	})
	log.Infof("DeleteObject")

	err := os.Remove(path)

	// This logic is specific to a file system; we need to clean up the backup directory
	// if there's nothing left. "Normal" object stores only mimic directory structures and don't need this.
	keyParts := strings.Split(key, "/")
	var backupPath string
	if len(keyParts) > 1 {
		backupPath = filepath.Join(getRoot(), bucket, keyParts[0], keyParts[1])
	}
	if backupPath != "" {
		infos, err := ioutil.ReadDir(backupPath)
		if err != nil {
			return err
		}
		if len(infos) == 0 {
			l := f.log.WithFields(logrus.Fields{
				"backupPath": backupPath,
			})
			l.Infof("Deleted backup directory")
			os.Remove(backupPath)
		}
	}

	return err
}

func (f *FileObjectStore) CreateSignedURL(bucket, key string, ttl time.Duration) (string, error) {
	log := f.log.WithFields(logrus.Fields{
		"bucket": bucket,
		"key":    key,
	})
	log.Infof("CreateSignedURL")
	return "", errors.New("CreateSignedURL is not supported for this plugin")
}

const defaultRoot = "/tmp/backups"

func getRoot() string {
	root := os.Getenv("ARK_FILE_OBJECT_STORE_ROOT")
	if root != "" {
		return root
	}

	return defaultRoot
}
