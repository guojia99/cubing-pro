package routes

import (
	"github.com/gin-gonic/gin"
	organizers2 "github.com/guojia99/cubing-pro/src/api/app/organizers"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

// 比赛和主办相关路由

func CompWithOrgRouters(router *gin.RouterGroup, svc *svc.Svc) {
	organizers := router.Group(
		"/organizers",
		middleware.JWT().MiddlewareFunc(),
		middleware.CheckAuthMiddlewareFunc(user.AuthPlayer), // 起码是一个玩家才能使用
	)
	{
		organizers.POST("/register", organizers2.RegisterOrganizers(svc))                         // 申请主办团队, 创建主办团队, 如果是管理员自动审核完成
		organizers.GET("/me", organizers2.MeOrganizers(svc))                                      // 我的主办团队列表, 包含申请中的
		organizers.GET("/:orgId", organizers2.OrgAuthMiddleware(svc), organizers2.Organizer(svc)) // 获取主办团队信息
	}

	person := organizers.Group(
		"/:orgId/person",
		middleware.CheckAuthMiddlewareFunc(user.AuthOrganizers),
		organizers2.OrgAuthMiddleware(svc),
	)
	{
		person.GET("/", organizers2.CheckOrgCanUse(), organizers2.Persons(svc))                   // 获取主办团队成员详细清单
		person.POST("/", organizers2.CheckOrgCanUse(), organizers2.CreatePersons(svc))            // 增加主办团队成员
		person.DELETE("/:personId", organizers2.CheckOrgCanUse(), organizers2.DeletePersons(svc)) // 删除主办团队成员
		person.DELETE("/exit", organizers2.Exit(svc))                                             // 退出主办团队
	}

	comp := organizers.Group(
		"/:orgId/comp",
		middleware.CheckAuthMiddlewareFunc(user.AuthOrganizers),
		organizers2.OrgAuthMiddleware(svc),
		organizers2.CheckOrgCanUse(),
	)
	{
		comp.GET("/", organizers2.OrgCompList(svc))             // 获取比赛列表
		comp.POST("/", organizers2.CreateComp(svc))             // 创建比赛 [需要提交审批]
		comp.GET("/:compId", organizers2.Comp(svc))             // 比赛详情
		comp.POST("/:compId/apply", organizers2.ApplyComp(svc)) // 申请比赛
		comp.DELETE("/:compId", organizers2.DeleteComp(svc))    // 删除比赛
		comp.POST("/:compId", organizers2.UpdateComp(svc))      // 更新比赛

		comp.GET("/:compId/players", organizers2.CompPlayers(svc))                  // 比赛选手列表 包含需审核
		comp.POST("/:compId/players/approval", organizers2.CompPlayerApproval(svc)) // 审核报名选手
		comp.POST("/:compId/players", organizers2.AddCompPlayer(svc))               // 添加比赛选手
		comp.DELETE("/:compId/players", organizers2.DeleteCompPlayer(svc))          // 移除比赛选手

		comp.POST("/:compId/result", organizers2.AddCompResult(svc))                           // 录入比赛成绩
		comp.DELETE("/:compId/result", organizers2.DeleteCompResult(svc))                      // 删除比赛成绩
		comp.GET("/:compId/pre_results", organizers2.GetCompPlayerPreResult(svc))              // 获取预录入成绩
		comp.POST("/:compId/pre_results/approval", organizers2.DeleteCompPlayerPreResult(svc)) // 审批预录入成绩
	}
}

func CompWithUserRouters(router *gin.RouterGroup, svc *svc.Svc) {
	userComp := router.Group(
		"/player_comp",
		middleware.CheckAuthMiddlewareFunc(user.AuthPlayer),
	)

	results := userComp.Group("/result")
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

	registers := userComp.Group("/register")
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
