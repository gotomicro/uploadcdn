package main

import (
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/server/egovernor"
	"uploadcdn/pkg/invoker"
	"uploadcdn/pkg/router"
)

func main() {
	err := ego.New().
		Invoker(
			invoker.Init,
		).
		Serve(
			egovernor.Load("server.governor").Build(),
			router.ServeHTTP(),
		).
		Run()
	if err != nil {
		elog.Panic("start up error: " + err.Error())
	}

}
