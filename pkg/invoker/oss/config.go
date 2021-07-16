package oss

import (
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
)

type Option func(c *Config)

// PackageName ..
const PackageName = "contrib.oss"

// Config ...
type Config struct {
	Debug           bool
	Mode            string
	Addr            string
	AccessKeyID     string
	AccessKeySecret string
	CdnName         string
	FileBucket      string
	logger          *elog.Component
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Debug:           false,
		Mode:            "oss",
		Addr:            "",
		AccessKeyID:     "",
		AccessKeySecret: "",
		CdnName:         "",
		FileBucket:      "",
		logger:          elog.EgoLogger.With(elog.FieldComponent(PackageName)),
	}
}

// Invoker ...
func Load(key string) *Config {
	var config = DefaultConfig()
	if err := econf.UnmarshalKey(key, &config); err != nil {
		config.logger.Panic("parse wechat config panic", elog.FieldErr(err), elog.FieldKey(key), elog.FieldValueAny(config))
	}
	config.logger = config.logger.With(elog.FieldComponentName(key))
	return config
}

// Build
func (cfg *Config) Build(options ...Option) *Component {
	for _, option := range options {
		option(cfg)
	}
	obj, err := NewComponent(cfg, cfg.Addr, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		cfg.logger.Panic("new component err", elog.FieldErr(err), elog.FieldValueAny(cfg))
	}
	return obj
}
