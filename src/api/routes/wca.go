package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/wca"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func WcaRouters(router *gin.RouterGroup, svc *svc.Svc) {

	w := router.Group("/wca")
	{
		w.GET("/player/:wcaID") // 选手信息
		w.GET("/player/:wcaID/rank_timers", wca.GetPersonRankTimer(svc))
	}

}
