package router

import (
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/server/egin"
	"uploadcdn/pkg/invoker"
)

func ServeHTTP() *egin.Component {
	r := egin.Load("server.http").Build()
	r.GET("/", func(ctx *gin.Context) {
		ctx.String(200, "Upload GoCN CDN")
	})
	r.POST("/upload", func(ctx *gin.Context) {
		//clientId := ctx.GetHeader("clientId")
		//clientSecret := ctx.GetHeader("clientSecret")
		//invoker.DB.

		var err error
		myfile, err := ctx.FormFile("myfile")
		if err != nil {
			ctx.String(400, err.Error())
			return
		}
		fileInfo, err := myfile.Open()
		if err != nil {
			ctx.String(400, err.Error())
			return
		}
		defer fileInfo.Close()

		err = invoker.Oss.PutObject(ctx.PostForm("name"), fileInfo)
		if err != nil {
			ctx.String(400, err.Error())
			return
		}
		ctx.String(200, "ok")
	})
	//r.POST("/test", func(ctx *gin.Context) {
	//	info, err := os.ReadFile("data/cloth.jpeg")
	//	if err != nil {
	//		ctx.String(400, err.Error())
	//		return
	//	}
	//	err = invoker.Oss.PutObject("test1/cloth.jpeg", bytes.NewReader(info))
	//	if err != nil {
	//		ctx.String(400, err.Error())
	//		return
	//	}
	//	ctx.String(200, "ok")
	//})
	return r
}
