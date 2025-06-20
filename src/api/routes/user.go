package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/users"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func UserRouters(router *gin.RouterGroup, svc *svc.Svc) {
	userGroup := router.Group("user",
		middleware.JWT().MiddlewareFunc(),
		middleware.CheckAuthMiddlewareFunc(user.AuthPlayer),
	)

	kv := userGroup.Group("kv")
	{
		kv.GET("/:key", users.GetKeyValue(svc))
		kv.POST("/", users.SetKeyValue(svc))
	}
}
