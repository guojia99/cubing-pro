package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/statics"
	"github.com/guojia99/cubing-pro/src/api/app/statistics"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func StaticRouters(router *gin.RouterGroup, svc *svc.Svc) {
	static := router.Group("static")
	{
		static.GET("/image/:uid", statics.Image(svc))
	}

	diyStatic := router.Group("diy_static")
	{
		diyStatic.GET("/diy_rankings/:key", statistics.DiyRankings(svc)) // 自定义版单

		diyStatic.GET("/diy_rankings", statistics.GetDiyRankingMaps(svc))                                                                                        // 获取所有列表
		diyStatic.POST("/diy_rankings", middleware.JWT().MiddlewareFunc(), middleware.CheckAuthMiddlewareFunc(user.AuthAdmin), statistics.AddDiyRankingMap(svc)) // 添加版单

		diyStatic.GET("/diy_rankings/:key/persons", middleware.JWT().MiddlewareFunc(), middleware.CheckAuthMiddlewareFunc(user.AuthAdmin), statistics.GetDiyRankingMapPersons(svc))
		diyStatic.POST("/diy_rankings/:key", middleware.JWT().MiddlewareFunc(), middleware.CheckAuthMiddlewareFunc(user.AuthAdmin), statistics.UpdateDiyRankingMapPersons(svc)) // 添加版单人、修改需要全部数据，相当于全量更新
	}
}
