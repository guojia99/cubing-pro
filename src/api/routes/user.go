package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/formulaforum"
	"github.com/guojia99/cubing-pro/src/api/app/notifications"
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

	forum := userGroup.Group("/formula-forum")
	forum.POST("/targets/:kind/:targetId/comments", formulaforum.CommentRateLimit(10, time.Minute), formulaforum.Create(svc.DB))
	forum.GET("/discussions", formulaforum.MyDiscussions(svc.DB))
	forum.GET("/comments", formulaforum.MyComments(svc.DB))
	forum.GET("/notifications", formulaforum.Notifications(svc.DB))
	forum.PUT("/notifications/:id/read", formulaforum.MarkRead(svc.DB))
	userGroup.GET("/notifications", notifications.List(svc.DB))
	userGroup.PUT("/notifications/:id/read", notifications.MarkRead(svc.DB))

	kv := userGroup.Group("kv")
	{
		kv.GET("", users.ListKeyValues(svc))
		kv.GET("/:key", users.GetKeyValue(svc))
		kv.POST("/", users.SetKeyValue(svc))
	}
}
