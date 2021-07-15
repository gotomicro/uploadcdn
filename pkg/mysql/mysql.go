package mysql

type CDN struct {
	ClientId       string // 发给用户的client id
	ClientSecret   string // 发给用户的client secret
	BucketName     string // 我们配置给用户的bucket name
	BucketClientId string // oss bucket id
}

func (*CDN) TableName() string {
	return "cdn"
}

type UploadLog struct {
}
