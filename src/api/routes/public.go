package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/guojia99/cubing-pro/src/svc"
)

func PublicRouters(router *gin.RouterGroup, svc *svc.Svc) {
	public := router.Group("/public")
	{
		public.GET("/swagger/json") // api文档
		public.GET("/events")       // 项目列表
		public.GET("/notify")       // 通知列表
		public.GET("/forum")        // 板块列表
	}

	player := public.Group("/player")
	{
		player.GET("/")                                     // 玩家列表
		player.GET("/search")                               // 搜索
		player.GET("/player/:playerId")                     // 玩家基础信息
		player.GET("/player/:playerId/month_report/:month") // 月度报表
		player.GET("/player/:playerId/year_report/:year")   // 年度报表
		player.GET("/player/:playerId/results")             // 玩家成绩汇总
		player.GET("/player/:playerId/nemesis")             // 宿敌列表
		player.GET("/player/:playerId/records")             // 玩家记录
		player.GET("/player/:playerId/sor")                 // 玩家统计成绩
	}

	comps := public.GET("/comps")
	{
		comps.GET("/")                  // 比赛列表
		comps.GET("/search")            // 搜索
		comps.GET("/:compId/registers") // 比赛报名列表
		comps.GET("/:compId/result")    // 比赛成绩列表
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
