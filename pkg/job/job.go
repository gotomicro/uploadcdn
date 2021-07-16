package job

import (
	"fmt"

	"github.com/gotomicro/ego/task/ejob"
	"uploadcdn/pkg/invoker"
	"uploadcdn/pkg/mysql"
)

func RunInstall(ctx ejob.Context) error {
	models := []interface{}{
		&mysql.CDN{},
	}
	err := invoker.DB.AutoMigrate(models...)
	if err != nil {
		fmt.Printf("err--------------->"+"%+v\n", err)
		return err
	}
	fmt.Println("create table ok")
	return nil
}
