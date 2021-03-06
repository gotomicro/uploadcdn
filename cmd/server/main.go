package main

import (
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/server/egovernor"
	"github.com/gotomicro/ego/task/ejob"
	"uploadcdn/pkg/invoker"
	"uploadcdn/pkg/job"
	"uploadcdn/pkg/router"
)

func main() {
	err := ego.New().
		Invoker(
			invoker.Init,
		).
		Job(
			ejob.Job("install", job.RunInstall),
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
