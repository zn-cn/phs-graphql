package token

import (
	"config"
	"fmt"

	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
)

var (
	accessKey = config.Conf.Qiniu.AccessKey
	secretKey = config.Conf.Qiniu.SecretKey
	bucket    = config.Conf.Qiniu.Bucket
)

func GetCustomUpToken(keyToOverwrite, persistentOps string, expires uint32) string {
	mac := qbox.NewMac(accessKey, secretKey)

	putPolicy := storage.PutPolicy{
		Scope:         fmt.Sprintf("%s:%s", bucket, keyToOverwrite),
		PersistentOps: persistentOps,
		Expires:       expires,
	}
	return putPolicy.UploadToken(mac)
}

func GetQiniuSimpleUpToken() string {
	// 简单上传凭证
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	mac := qbox.NewMac(accessKey, secretKey)
	return putPolicy.UploadToken(mac)

}

func GetQiniuSimpleUpTokenWithExpires() string {
	// 设置上传凭证有效期
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	putPolicy.Expires = 7200 //示例2小时有效期
	mac := qbox.NewMac(accessKey, secretKey)
	return putPolicy.UploadToken(mac)
}

// GetQiniuOverwriteUpToken keyToOverwrite  需要覆盖的文件名
func GetQiniuOverwriteUpToken(keyToOverwrite string) string {
	// 覆盖上传凭证
	putPolicy := storage.PutPolicy{
		Scope: fmt.Sprintf("%s:%s", bucket, keyToOverwrite),
	}
	mac := qbox.NewMac(accessKey, secretKey)
	return putPolicy.UploadToken(mac)
}

func GetQiniuCustomRetUpToken() string {
	// 自定义上传回复凭证
	mac := qbox.NewMac(accessKey, secretKey)

	putPolicy := storage.PutPolicy{
		Scope:      bucket,
		ReturnBody: `{"key":"$(key)","hash":"$(etag)","fsize":$(fsize),"bucket":"$(bucket)","name":"$(x:name)"}`,
	}
	return putPolicy.UploadToken(mac)
}

// GetQiniuJSONCallbackUpToken callbackURL: "http://api.example.com/qiniu/upload/callback"
func GetQiniuJSONCallbackUpToken(callbackURL string) string {
	// 带回调业务服务器的凭证(JSON方式)
	mac := qbox.NewMac(accessKey, secretKey)
	putPolicy := storage.PutPolicy{
		Scope:            bucket,
		CallbackURL:      callbackURL,
		CallbackBody:     `{"key":"$(key)","hash":"$(etag)","fsize":$(fsize),"bucket":"$(bucket)","name":"$(x:name)"}`,
		CallbackBodyType: "application/json",
	}
	return putPolicy.UploadToken(mac)
}

// GetQiniuURLCallbackUpToken callbackURL: "http://api.example.com/qiniu/upload/callback"
func GetQiniuURLCallbackUpToken(callbackURL string) string {
	// 带回调业务服务器的凭证（URL方式）
	mac := qbox.NewMac(accessKey, secretKey)

	putPolicy := storage.PutPolicy{
		Scope:        bucket,
		CallbackURL:  callbackURL,
		CallbackBody: "key=$(key)&hash=$(etag)&bucket=$(bucket)&fsize=$(fsize)&name=$(x:name)",
	}
	return putPolicy.UploadToken(mac)
}

// GetQiniuOpsUpToken persistentNotifyURL: "http://api.example.com/qiniu/pfop/notify"
/*
	// 带数据处理的凭证
	saveMp4Entry := base64.URLEncoding.EncodeToString([]byte(bucket + ":avthumb_test_target.mp4"))
	saveJpgEntry := base64.URLEncoding.EncodeToString([]byte(bucket + ":vframe_test_target.jpg"))
	//数据处理指令，支持多个指令
	avthumbMp4Fop := "avthumb/mp4|saveas/" + saveMp4Entry
	vframeJpgFop := "vframe/jpg/offset/1|saveas/" + saveJpgEntry
	//连接多个操作指令
	persistentOps := strings.Join([]string{avthumbMp4Fop, vframeJpgFop}, ";")
*/
func GetQiniuOpsUpToken(persistentOps, persistentNotifyURL string) string {
	mac := qbox.NewMac(accessKey, secretKey)

	putPolicy := storage.PutPolicy{
		Scope:               bucket,
		PersistentOps:       persistentOps,
		PersistentNotifyURL: persistentNotifyURL,
	}
	return putPolicy.UploadToken(mac)
}
