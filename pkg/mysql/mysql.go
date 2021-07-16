package mysql

type CDN struct {
	ID             int    // id 号
	ClientId       string // 发给用户的client id
	ClientSecret   string // 发给用户的client secret
	BucketDir      string // 我们配置给用户的bucket dir
	BucketName     string // 与client id对应的bucket name
	BucketClientId string // oss bucket id
}

func (*CDN) TableName() string {
	return "cdn"
}

type UploadLog struct {
}
