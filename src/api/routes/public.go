package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/comp"
	"github.com/guojia99/cubing-pro/src/api/app/notify"
	posts "github.com/guojia99/cubing-pro/src/api/app/post"
	"github.com/guojia99/cubing-pro/src/api/app/result"
	"github.com/guojia99/cubing-pro/src/api/app/users"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func PublicRouters(router *gin.RouterGroup, svc *svc.Svc) {
	public := router.Group("/public")
	{
		public.GET("/swagger/json", func(context *gin.Context) { context.JSON(404, "not data") }) // api文档
		public.GET("/events", result.Events(svc))                                                 // 项目列表 TODO 加缓存
		public.GET("/notify", notify.GetNotifyList(svc))                                          // 通知列表
		public.GET("/forum", posts.GetForums(svc))                                                // 板块列表
	}

	player := public.Group("/player")
	{
		// todo 加缓存
		player.GET("/", users.Users(svc))                          // 玩家列表
		player.GET("/player/:playerId", users.UserBaseResult(svc)) // 玩家基础信息
		//player.GET("/player/:playerId/report/", result.PlayerTimeReports(svc)) // 报表
		player.GET("/player/:playerId/results", result.PlayerResults(svc)) // 玩家成绩汇总
		player.GET("/player/:playerId/nemesis", result.PlayerNemesis(svc)) // 宿敌列表
		player.GET("/player/:playerId/records", result.PlayerRecords(svc)) // 玩家记录
		player.GET("/player/:playerId/sor", result.PlayerSor(svc))         // 玩家统计成绩
	}

	comps := public.Group("/comps")
	{
		comps.GET("/", comp.List(svc))                       // 比赛列表
		comps.POST("/", comp.List(svc))                      // 查询
		comps.GET("/:compId", comp.Comp(svc))                // 比赛详情
		comps.GET("/:compId/registers", comp.Registers(svc)) // 比赛报名列表
		comps.GET("/:compId/result", comp.Results(svc))      // 比赛成绩列表
		comps.GET("/:compId/record", comp.Record(svc))       // 比赛产生的记录
	}

	statistics := public.Group("/statistics")
	{
		statistics.GET("/best_result")            //最佳成绩列表
		statistics.GET("/records")                //记录列表
		statistics.GET("/sor")                    //Sor统计
		statistics.GET("/sum-of-ranks")           //排名总和榜单,可分单次平均、选择项目
		statistics.GET("/medal-collection")       //奖牌累积榜单，可分项目
		statistics.GET("/top-n")                  //项目前N - 指该项目前N的历史成绩（不根据选手去重，选手可以重复上榜），可分单平
		statistics.GET("/record-num")             //记录数
		statistics.GET("/comp-record-num")        //赛事打破记录数
		statistics.GET("/record-time")            //记录保持时间榜单
		statistics.GET("/most-comps-num")         //选手比赛记录数
		statistics.GET("/most-persons-in-comps")  //赛事人数排名
		statistics.GET("/most-solves-by-persons") //选手还原次数排名
		statistics.GET("/most-solves-in-comps")   //赛事还原次数排名
		statistics.GET("/most-personal-solves")   //选手还原次数排名，可按年份区分
		statistics.GET("/best-uncrowned-kings")   //无冕之王, 排在第二里面成绩最好
		statistics.GET("/best-podium-miss")       //老四之王，排在第四里面成绩最好
		statistics.GET("/all-events")             //大满贯
	}
}
