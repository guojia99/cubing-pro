package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func CompRouters(router *gin.RouterGroup, svc *svc.Svc) {
	organizers := router.Group("/organizers")
	{
		organizers.GET("/:organizersId")              //获取主办团队信息
		organizers.POST("/")                          //创建主办团队
		organizers.POST("/add_person/:personId")      //增加主办团队成员
		organizers.DELETE("/delete_person/:personId") //删除主办团队成员
		organizers.DELETE("/exit")                    //退出主办团队
		organizers.GET("/me")                         //我的主办团队列表

		// 管理员
		organizers.GET("/")                 //主办团队列表
		organizers.DELETE("/:organizersId") //删除主办团队
	}

	comp := router.Group("/competition")
	{
		comp.GET("/")                    // 获取比赛列表
		comp.POST("/")                   // 创建比赛
		comp.GET("/:compId")             // 比赛详情
		comp.GET("/me")                  // 获取我管理的比赛
		comp.POST("/:compId/apply")      // 申请比赛
		comp.DELETE("/:compId")          // 删除比赛
		comp.POST("/:compId")            // 更新比赛
		comp.POST("/:compId/add_player") // 添加比赛选手

		// 管理员
		comp.GET("/approvals/comps")             // 比赛审批列表
		comp.POST("/approvals/:compId/approval") // 比赛审批
	}

	results := router.Group("/result")
	{
		results.POST("/:compId")   // 录入比赛成绩
		results.DELETE("/:compId") // 删除比赛成绩

		results.GET("/:compId/pre_results")           // 【主办】获取预录入成绩列表
		results.POST("/:compId/approval/pre_results") // 【主办】审批预录入成绩
		results.POST("/:compId/add_results")          // 预录入比赛成绩
		results.GET("/pre_results/me")                // 获取我的预录入列表
		results.POST("/add/:compId")                  // 添加预录入成绩
		results.DELETE("/delete/:pre_id")             // 删除预录入成绩
	}

	registers := router.Group("/register")
	{
		registers.GET("/comps")                  // 报名比赛列表
		registers.GET("/comps/:compId/detail")   // 报名详情
		registers.POST("/comps/:compId/")        // 报名比赛
		registers.GET("/comps/:compId/callback") // 报名比赛支付回调
		registers.GET("/comps/:compId/progress") // 报名比赛支付进度查询
		registers.PUT("/comps/:compId/events")   // 添加比赛项目
		registers.POST("/comps/:compId/retire")  // 退赛
	}
}
