package routes

import (
	"github.com/gin-gonic/gin"
	events2 "github.com/guojia99/cubing-pro/src/api/app/events"
	notify3 "github.com/guojia99/cubing-pro/src/api/app/notify"
	"github.com/guojia99/cubing-pro/src/api/app/organizers"
	posts "github.com/guojia99/cubing-pro/src/api/app/post"
	systemResults "github.com/guojia99/cubing-pro/src/api/app/systemResult"
	"github.com/guojia99/cubing-pro/src/api/app/users"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	user2 "github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func AdminRouters(router *gin.RouterGroup, svc *svc.Svc) {
	admin := router.Group(
		"/admin",
		middleware.JWT().MiddlewareFunc(),
		middleware.CheckAuthMiddlewareFunc(user2.AuthSuperAdmin),
	)

	//// 角色管理
	//userRole := admin.Group("/user_role").Use(middleware.JWT().MiddlewareFunc()) // todo 权限控制
	//{
	//	userRole.GET("/")                         // 角色列表
	//	userRole.POST("/")                        // 添加角色
	//	userRole.DELETE("/")                      // 删除角色
	//	userRole.PUT("/")                         // 修改角色
	//	userRole.GET("/:roleId/users")            // 角色对应用户列表
	//	userRole.PUT("/:roleId/bind")             // 用户绑定角色
	//	userRole.PUT("/:roleId/unbind")           // 用户解绑角色
	//	userRole.GET("/:roleId/rules")            // 角色权限列表
	//	userRole.POST("/:roleId/rules")           // 新增角色权限
	//	userRole.DELETE("/:roleId/rules/:ruleId") // 删除角色权限
	//	userRole.PUT("/:roleId/rules/:ruleId")    // 修改角色权限
	//}

	// 项目管理
	event := admin.Group("/events")
	{
		event.GET("/", events2.Events(svc))         // 项目列表
		event.POST("/", events2.CreateEvents(svc))  // 新增项目
		event.DELETE("/", events2.DeleteEvent(svc)) // 移除项目
	}

	// 通知管理
	notify := admin.Group("/notify")
	{
		notify.GET("/", notify3.List(svc))                     // 通知列表
		notify.GET("/:notifyId", notify3.NotifyDetail(svc))    // 通知详情
		notify.POST("/", notify3.CreateNotify(svc))            // 添加通知
		notify.PUT("/:notifyId", notify3.UpdateNotify(svc))    // 修改通知
		notify.DELETE("/:notifyId", notify3.DeleteNotify(svc)) // 删除通知
	}

	// 系统配置
	systemResult := admin.Group("/system_result")
	{
		systemResult.GET("/", systemResults.GetSystemResult(svc))         // 获取系统相关配置
		systemResult.PUT("/title", systemResults.SetSystemTitle(svc))     // 设置网站标题
		systemResult.PUT("/welcome", systemResults.SetSystemWelcome(svc)) // 设置欢迎词
		systemResult.PUT("/footer", systemResults.SetSystemFooter(svc))   // 设置网站脚注
		systemResult.PUT("/logo", systemResults.SetSystemLogo(svc))       // 设置网站logo
		systemResult.PUT("/:key", systemResults.SetSystemKeyValue(svc))   // 设置系统配置 key value
	}

	// 帖子管理
	post := admin.Group("/post")
	{
		post.GET("/forums", posts.GetForums(svc))              // 板块列表
		post.POST("/forum", posts.CreateForum(svc))            // 添加板块
		post.DELETE("/forum/:forumId", posts.DeleteForum(svc)) // 删除板块

		post.GET("/topics", posts.GetAllTopics(svc))            // 获取所有帖子(包括被删的 \ 禁用的)
		post.GET("/topics/:topicId", posts.GetTopic(svc, true)) // 帖子详情
		post.DELETE("/topics/:topicId", posts.DeleteTopic(svc)) // 删除帖子
		post.PUT("/topics/ban/:postId", posts.BanTopic(svc))    // 禁用帖子

		post.GET("/topics/:topicId/posts", posts.GetPosts(svc, true))        // 获取帖子评论列表
		post.DELETE("/topics/:topicId/posts/:postId", posts.DeletePost(svc)) // 删除帖子的评论
	}

	// 用户管理
	user := admin.Group("/user")
	{
		user.PUT("/ban", users.BanUser(svc))                              // 禁用用户
		user.PUT("/reset_password", users.RetrievePasswordWithAdmin(svc)) // 授权重置用户密码
	}

	// 主办团队
	comp := admin.Group("/competition")
	{
		// 管理员
		comp.GET("/organizers", organizers.AllOrganizers(svc))                 //主办团队列表
		comp.POST("/:orgId", organizers.DoWithOrganizers(svc))                 // 处理主办团队， 禁用等
		comp.GET("/approvals/comps", organizers.Comps(svc))                    // 比赛审批列表
		comp.POST("/approvals/:compId/approval", organizers.ApprovalComp(svc)) // 比赛审批
	}
}
