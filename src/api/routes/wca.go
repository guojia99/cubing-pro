package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/wca"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func WcaRouters(router *gin.RouterGroup, svc *svc.Svc) {

	w := router.Group("/wca")
	{
		w.GET("/players/:name", wca.Players(svc))

		w.GET("/player/:wcaID", wca.PlayerPersonInfo(svc)) // 选手信息
		w.GET("/player/:wcaID/results", wca.PlayerResults(svc))
		w.GET("/player/:wcaID/competitions", wca.PlayerCompetitions(svc))
		w.GET("/player/:wcaID/rank_timers", wca.GetPersonRankTimer(svc))

		w.GET("/country", wca.Country(svc))
		w.POST("/ranks/:eventID", wca.GetEventRankWithTimer(svc))
	}
}
