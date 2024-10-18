package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/comp"
	"github.com/guojia99/cubing-pro/src/api/app/notify"
	posts "github.com/guojia99/cubing-pro/src/api/app/post"
	"github.com/guojia99/cubing-pro/src/api/app/result"
	"github.com/guojia99/cubing-pro/src/api/app/statistics"
	"github.com/guojia99/cubing-pro/src/api/app/users"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"time"
)

func PublicRouters(router *gin.RouterGroup, svc *svc.Svc) {
	public := router.Group("/public",
		middleware.CacheMiddleware(time.Minute*5),
		middleware.RateLimitMiddleware(20, time.Second))
	{
		public.GET("/swagger/json", func(context *gin.Context) { context.JSON(404, "not data") }) // api文档
		public.GET("/events", result.Events(svc))                                                 // 项目列表 TODO 加缓存
		public.GET("/notify", notify.GetNotifyList(svc))                                          // 通知列表
		public.GET("/forum", posts.GetForums(svc))                                                // 板块列表
	}

	player := public.Group("/player", middleware.CacheMiddleware(time.Minute))
	{
		player.Any("/", users.Users(svc))                 // 玩家列表
		player.GET("/:cubeId", users.UserBaseResult(svc)) // 玩家基础信息
		//player.GET("/player/:playerId/report/", result.PlayerTimeReports(svc)) // 报表
		player.GET("/:cubeId/results", result.PlayerResults(svc)) // 玩家成绩汇总
		player.GET("/:cubeId/nemesis", result.PlayerNemesis(svc)) // 宿敌列表
		player.GET("/:cubeId/records", result.PlayerRecords(svc)) // 玩家记录
		player.GET("/:cubeId/sor", result.PlayerSor(svc))         // 玩家统计成绩
		player.GET("/:cubeId/comps", result.PlayerComps(svc))     // 玩家参加过的比赛列表
	}

	comps := public.Group("/comps")
	{
		comps.Any("/", comp.List(svc))                       // 比赛列表 查询
		comps.GET("/:compId", comp.Comp(svc))                // 比赛详情
		comps.GET("/:compId/registers", comp.Registers(svc)) // 比赛报名列表
		comps.GET("/:compId/result", comp.Results(svc))      // 比赛成绩列表
		comps.GET("/:compId/record", comp.Record(svc))       // 比赛产生的记录
	}

	sta := public.Group("/statistics")
	{
		sta.Any("/best_result", statistics.Best(svc))          //最佳成绩列表
		sta.Any("/best_result/:eventId", statistics.Best(svc)) //最佳成绩基于项目单项
		sta.Any("/records", statistics.Records(svc))           //记录列表
		sta.Any("/kinch", statistics.KinCh(svc))               //Sor统计
		sta.GET("/sum-of-ranks")                               //排名总和榜单,可分单次平均、选择项目
		sta.GET("/medal-collection")                           //奖牌累积榜单，可分项目
		sta.GET("/top-n")                                      //项目前N - 指该项目前N的历史成绩（不根据选手去重，选手可以重复上榜），可分单平
		sta.GET("/record-num")                                 //记录数
		sta.GET("/comp-record-num")                            //赛事打破记录数
		sta.GET("/record-time")                                //记录保持时间榜单
		sta.GET("/most-comps-num")                             //选手比赛记录数
		sta.GET("/most-persons-in-comps")                      //赛事人数排名
		sta.GET("/most-solves-by-persons")                     //选手还原次数排名
		sta.GET("/most-solves-in-comps")                       //赛事还原次数排名
		sta.GET("/most-personal-solves")                       //选手还原次数排名，可按年份区分
		sta.GET("/best-uncrowned-kings")                       //无冕之王, 排在第二里面成绩最好
		sta.GET("/best-podium-miss")                           //老四之王，排在第四里面成绩最好
		sta.GET("/all-events")                                 //大满贯
	}
}
