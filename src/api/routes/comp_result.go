package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/comp"
	organizers2 "github.com/guojia99/cubing-pro/src/api/app/organizers"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/app/result"
	"github.com/guojia99/cubing-pro/src/api/app/users"
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
		//middleware.RateLimitMiddleware(20, time.Second),
	)
	{
		organizers.POST("/register", organizers2.RegisterOrganizers(svc))                            // 申请主办团队, 创建主办团队, 如果是管理员自动审核完成
		organizers.GET("/me", organizers2.MeOrganizers(svc))                                         // 我的主办团队列表, 包含申请中的
		organizers.GET("/:orgId", org_mid.OrgAuthMiddleware(svc), organizers2.Organizer(svc))        // 获取主办团队信息
		organizers.GET("/:orgId/groups", org_mid.OrgAuthMiddleware(svc), organizers2.GetGroups(svc)) // 获取群组
	}

	person := organizers.Group(
		"/:orgId/person",
		middleware.CheckAuthMiddlewareFunc(user.AuthOrganizers),
		org_mid.OrgAuthMiddleware(svc),
	)
	{
		person.GET("/", org_mid.CheckOrgCanUse(), organizers2.Persons(svc))                   // 获取主办团队成员详细清单
		person.POST("/", org_mid.CheckOrgCanUse(), organizers2.CreatePersons(svc))            // 增加主办团队成员
		person.DELETE("/:personId", org_mid.CheckOrgCanUse(), organizers2.DeletePersons(svc)) // 删除主办团队成员
		person.DELETE("/exit", organizers2.ExitOrganizer(svc))                                // 退出主办团队
	}

	compR := organizers.Group(
		"/:orgId/comp",
		middleware.CheckAuthMiddlewareFunc(user.AuthOrganizers),
		org_mid.OrgAuthMiddleware(svc),
		org_mid.CheckOrgCanUse(),
	)
	{

		compR.GET("/", organizers2.OrgCompList(svc)) // 获取比赛列表
		compR.POST("/", organizers2.CreateComp(svc)) // 创建比赛 [需要提交审批]

		compId := compR.Group(
			"/:compId",
			org_mid.CheckCompMiddleware(svc),
		)
		{
			compId.GET("", organizers2.Comp(svc))             // 比赛详情
			compId.POST("/apply", organizers2.ApplyComp(svc)) // 申请比赛
			compId.DELETE("", organizers2.DeleteComp(svc))    // 删除比赛
			compId.POST("", organizers2.UpdateComp(svc))      // 更新比赛
			compId.POST("/end", organizers2.EndComp(svc))     // 结束比赛

			compId.GET("/all_players", users.Users(svc, 0))                               // 临时API， 用于获取所有的选手
			compId.GET("/players", organizers2.CompPlayers(svc))                          // 比赛选手列表 包含需审核
			compId.POST("/players/approval/:reg_id", organizers2.CompPlayerApproval(svc)) // 审核报名选手

			//compId.POST("/players", organizers2.AddCompPlayer(svc))                       // 添加比赛选手
			//compId.DELETE("/players", organizers2.DeleteCompPlayer(svc))                  // 移除比赛选手

			compId.GET("/result", organizers2.GetCompResult(svc))
			compId.POST("/result", organizers2.AddCompResult(svc))                                        // 录入比赛成绩
			compId.DELETE("/result/:result_id", organizers2.DeleteCompResult(svc))                        // 删除比赛成绩
			compId.GET("/pre_results", organizers2.GetCompPlayerPreResult(svc))                           // 获取预录入成绩
			compId.POST("/pre_results/:result_id/approval", organizers2.ApprovalCompPlayerPreResult(svc)) // 审批预录入成绩

			compId.PUT("/:reg_id/:compId/refresh_event", organizers2.RefreshEvent(svc)) // 刷新项目的轮次信息
		}

	}
}

func CompWithUserRouters(router *gin.RouterGroup, svc *svc.Svc) {
	userComp := router.Group(
		"/player_comp",
		middleware.CheckAuthMiddlewareFunc(user.AuthPlayer),
	)

	results := userComp.Group("/result")
	{
		results.GET("/pre/list", result.PreResults(svc))                   // 获取我的预录入列表
		results.POST("/pre/add_results", result.AddPreResults(svc))        // 预录入比赛成绩
		results.DELETE("/pre/delete/:pre_id", result.DeletePreResult(svc)) // 删除预录入成绩
	}

	router.GET("/player_comp/register/comps/:compId/callback/:registerId", comp.RegisterCompCallback(svc)) // 报名比赛支付回调
	registers := userComp.Group(
		"/register",
	)
	{
		registers.GET("/comps", comp.RegisterComps(svc))                     // 报名比赛列表
		registers.POST("/comps/:compId/", comp.RegisterComp(svc))            // 报名比赛
		registers.GET("/comps/:compId/detail", comp.RegisterCompDetail(svc)) // 报名详情
		//registers.GET("/comps/:compId/progress", comp.RegisterProgress(svc))   // 报名比赛支付进度查询
		registers.PUT("/comps/:compId/add_events", comp.RegisterAddEvent(svc)) // 添加比赛项目
		registers.POST("/comps/:compId/retire", comp.RegisterRetire(svc))      // 退赛
	}

}
