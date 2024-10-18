package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/statics"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func StaticRouters(router *gin.RouterGroup, svc *svc.Svc) {
	static := router.Group("static")
	{
		static.GET("/image/:uid", statics.Image(svc))
	}
}
