package oss

import (
	"fmt"
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type Component struct {
	*oss.Client
	config *Config
}

func NewComponent(config *Config, endpoints, accessKeyId, accessKeySecret string) (client *Component, err error) {
	c, e := oss.New(
		endpoints, accessKeyId, accessKeySecret,
	)
	if e != nil {
		return
	}

	client = &Component{
		Client: c,
		config: config,
	}
	return
}

func (c *Component) PutObject(bucketName string, dstPath string, reader io.Reader) error {
	b, e := c.Bucket(bucketName)
	if e != nil {
		return fmt.Errorf("get bucket failed, err: %w", e)
	}
	return b.PutObject(dstPath, reader)
}
