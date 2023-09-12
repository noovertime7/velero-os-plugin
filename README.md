# Cloud Velero Object Store Plugins

## 背景
公司备份平台深度使用velero作为备份工具，但是最新版本aws的备份插件在上传至阿里云OSS中报不支持的错误，使用阿里OSS插件无法备份到minio中，所以自研开发了这款备份插件
。

## 功能
- 支持阿里OSS上传备份文件
- 支持minio上传备份文件
- 更多... 欢迎PR

## 原理
分别使用minio sdk与阿里oss sdk实现文件上传，实现了velero的Object Store接口，供velero调用


## 如何使用
- velero注册插件
如果已经安装了velero，注册将变得非常简单，命令如下：
```shell
 velero plugin add <Image>:<Version>
```
执行命令后会在velero容器内添加一个初始化容器

如果并未安装velero，请参阅我的blog ```yunxue521.top/velero```


- 修改```backupstoragelocation```CRD的provider为cloud

```shell
 kubectl edit backupstoragelocation -n <VeleroNamespace> <Name>

config:
    profile: oss
    region: beijing
    s3ForcePathStyle: "false"
    s3Type: oss
    s3Url: <OssEndpoint>
  credential:
    key: cloud
    name: <secretName>
  objectStorage:
    bucket: <Bucket>
  provider: cloud
```
重要的是provider要修改为cloud，即可正常使用，如果是minio需要将s3Type修改为minio
