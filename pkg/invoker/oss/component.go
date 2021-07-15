package oss

import (
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type Component struct {
	b      *oss.Bucket
	config *Config
}

func NewComponent(config *Config, endpoints, accessKeyId, accessKeySecret, bucketName string, isDelete bool) (client *Component, err error) {
	c, e := oss.New(
		endpoints, accessKeyId, accessKeySecret,
	)
	if e != nil {
		return
	}
	b, e := c.Bucket(bucketName)
	if e != nil {
		return
	}
	client = &Component{
		config: config,
		b:      b,
	}
	return
}

func (c *Component) PutObject(dstPath string, reader io.Reader) error {
	return c.b.PutObject(dstPath, reader)
}
