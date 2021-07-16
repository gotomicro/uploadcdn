package invoker

import (
	"github.com/gotomicro/ego-component/egorm"
	"uploadcdn/pkg/invoker/oss"
)

var (
	Oss *oss.Component
	DB  *egorm.Component
)

func Init() error {
	Oss = oss.Load("oss").Build()
	DB = egorm.Load("mysql").Build()
	return nil
}
