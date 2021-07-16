package router

import (
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/server/egin"
	"uploadcdn/pkg/invoker"
	"uploadcdn/pkg/mysql"
)

func ServeHTTP() *egin.Component {
	r := egin.Load("server.http").Build()
	r.GET("/", func(ctx *gin.Context) {
		ctx.String(200, "Upload CDN")
	})
	r.POST("/upload", func(ctx *gin.Context) {
		clientId := ctx.GetHeader("clientId")
		clientSecret := ctx.GetHeader("clientSecret")

		var cdnInfo mysql.CDN
		invoker.DB.Where("client_id = ? and client_secret = ?", clientId, clientSecret).Find(&cdnInfo)
		if cdnInfo.ID == 0 {
			ctx.String(400, "没有找到信息")
			return
		}

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

		err = invoker.Oss.PutObject(cdnInfo.BucketName, cdnInfo.BucketDir+"/"+ctx.PostForm("name"), fileInfo)
		if err != nil {
			ctx.String(400, err.Error())
			return
		}
		ctx.String(200, "ok")
	})
	return r
}
