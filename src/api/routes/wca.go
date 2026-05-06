package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/wca"
	"github.com/guojia99/cubing-pro/src/api/middleware"
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
		w.POST("/ranks/all-events-achiever", wca.BaseStaticsWithEventAndCacheKey(svc, "GetAllEventsAchievement"))
		w.POST("/ranks/diy_events", wca.BaseStaticsWithEventAndCacheKey(svc, "GetRankWithEvents"))
		w.POST("/rank/rank-with-start-comp-year/:eventID", wca.BaseStaticsWithEventAndCacheKey(svc, "GetWithCompYearPersonRank")) // 参赛年限的人总排行

		w.GET("/grand-slam", wca.GetGrandSlam(svc))

		// 粗饼选手主页代理（服务端抓取 HTML）；每 IP 限流；出站串行+节流在 Handler 内（见 cubing_china_person）
		w.GET("/cubing-china/person/:wcaID", middleware.RateLimitMiddleware(80, time.Minute), wca.CubingChinaPerson(svc))
	}

	extendStatic := w.Group("/extend")
	{
		extendStatic.GET("/resultProportionEstimation", wca.ResultProportionEstimation(svc)) // 成绩拟合
	}
}
