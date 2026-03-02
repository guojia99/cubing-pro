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

		w.GET("/player/:wcaID", wca.BaseGetPlayerWithKey(svc, "GetPersonInfo")) // 选手信息
		w.GET("/player/:wcaID/results", wca.PlayerResults(svc))
		w.GET("/player/:wcaID/competitions", wca.PlayerCompetitions(svc))
		w.GET("/player/:wcaID/rank_timers", wca.GetPersonRankTimer(svc))
		w.GET("/player/:wcaID/best_ranks", wca.BaseGetPlayerWithKey(svc, "GetPersonBestRanks"))

		w.GET("/country", wca.Country(svc))
		w.POST("/ranks/historical/full/:eventID", wca.BaseStaticsWithEventAndCacheKey(svc, "GetEventRankWithTimer")) // 截止某年
		w.POST("/ranks/full/:eventID", wca.BaseStaticsWithEventAndCacheKey(svc, "GetEventRankWithFullNow"))
		w.POST("/ranks/historical/:eventID", wca.BaseStaticsWithEventAndCacheKey(svc, "GetEventRankWithOnlyYear")) // 全年排名
		w.POST("/ranks/success_rate/:eventID", wca.BaseStaticsWithEventAndCacheKey(svc, "GetEventSuccessRateResult"))
	}
}
