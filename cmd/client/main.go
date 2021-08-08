package main

import (
	"strings"

	"github.com/gotomicro/ego/core/eflag"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/util/xtime"
	"uploadcdn/cmd/client/upload"
)

func init() {

	eflag.Register(&eflag.StringFlag{
		Name:  "dir",
		Usage: "--dir",
	})

	eflag.Register(&eflag.StringFlag{
		Name:    "id",
		Usage:   "--id",
		Default: "",
	})
	eflag.Register(&eflag.StringFlag{
		Name:    "secret",
		Usage:   "--secret",
		Default: "",
	})

	eflag.Register(&eflag.StringFlag{
		Name:    "addr",
		Usage:   "--addr",
		Default: "https://upload.gocn.vip",
	})
	eflag.Register(&eflag.BoolFlag{
		Name:    "debug",
		Usage:   "--debug",
		Default: false,
	})
	eflag.Register(&eflag.StringFlag{
		Name:    "timeout",
		Usage:   "--timeout",
		Default: "10s",
	})
}

func main() {
	eflag.Parse()
	dir := eflag.String("dir")
	var dirs []string
	if strings.Contains(dir, ",") {
		dirs = strings.Split(dir, ",")
	} else {
		dirs = []string{dir}
	}

	err := upload.RunCommand(
		eflag.String("id"),
		eflag.String("secret"),
		eflag.String("addr"),
		dirs,
		eflag.Bool("debug"),
		xtime.Duration(eflag.String("timeout")),
	)
	if err != nil {
		elog.Panic("upload panic", elog.FieldErr(err))
	}
}
