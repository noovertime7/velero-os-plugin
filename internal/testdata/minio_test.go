package testdata

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/velero-plugin-example/internal/plugin/uploader"
	"path"
	"strings"
	"testing"
)

func TestMinioUploader_ListCommonPrefixes(t *testing.T) {
	upload, err := uploader.NewMinioUploader("http://backup-devops.tj-it.com.cn", "admin", "08b72d98eb7daa56180c3d8d7f41e62f", false, "minio", logrus.New())
	if err != nil {
		t.Fatal(err)
	}
	data, err := upload.ListCommonPrefixes("test", "backups", "/")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(data)
	for _, dir := range data {
		subdir := strings.TrimSuffix(strings.TrimPrefix(dir, ""), "/")
		fmt.Println(subdir)
	}
	fmt.Println("", path.Join("", "backups")+"/")
}

func TestMinioUploader_ListObjects(t *testing.T) {
	upload, err := uploader.NewMinioUploader("http://backup-devops.tj-it.com.cn", "admin", "08b72d98eb7daa56180c3d8d7f41e62f", false, "minio", logrus.New())
	if err != nil {
		t.Fatal(err)
	}
	data, err := upload.ListObjects("test", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(data)
}

func TestMinioUploader_GetObject(t *testing.T) {
	upload, err := uploader.NewMinioUploader("http://192.168.11.207:9000", "admin", "Tsit@2022", false, "minio", logrus.New())
	if err != nil {
		t.Fatal(err)
	}
	data, err := upload.GetObject("alert", "images/0ca52ytp6QTCIZCyFjSt.png")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(data)
}

func TestMinioUploader_DeleteObject(t *testing.T) {
	upload, err := uploader.NewMinioUploader("http://backup-devops.tj-it.com.cn", "admin", "08b72d98eb7daa56180c3d8d7f41e62f", false, "minio", logrus.New())
	if err != nil {
		t.Fatal(err)
	}
	err = upload.DeleteObject("alert", "images/0ca52ytp6QTCIZCyFjSt.png")
	if err != nil {
		t.Fatal(err)
	}
}
