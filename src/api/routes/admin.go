package routes

import (
	"github.com/gin-gonic/gin"
	events2 "github.com/guojia99/cubing-pro/src/api/app/events"
	notify3 "github.com/guojia99/cubing-pro/src/api/app/notify"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	user2 "github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func AdminRouters(router *gin.RouterGroup, svc *svc.Svc) {
	admin := router.Group("/admin")

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
	event := admin.Group("/events").Use(
		middleware.JWT().MiddlewareFunc(),
		middleware.CheckAuthMiddlewareFunc(user2.AuthAdmin),
	)
	{
		event.GET("/", events2.Events(svc))         // 项目列表
		event.POST("/", events2.CreateEvents(svc))  // 新增项目
		event.DELETE("/", events2.DeleteEvent(svc)) // 移除项目
	}

	// 通知管理
	notify := admin.Group("/notify").Use(middleware.JWT().MiddlewareFunc(), middleware.CheckAuthMiddlewareFunc(user2.AuthAdmin)) // todo 权限控制
	{
		notify.GET("/", notify3.List(svc))                  // 通知列表
		notify.GET("/:notifyId", notify3.NotifyDetail(svc)) // 通知详情
		notify.POST("/", notify3.CreateNotify(svc))         // 添加通知
		notify.PUT("/:notifyId", notify3.UpdateNotify(svc)) // 修改通知
		notify.DELETE("/", notify3.DeleteNotify(svc))       // 删除通知
	}

	// 系统配置
	systemResult := admin.Group("/system_result").Use(middleware.JWT().MiddlewareFunc()) // todo 权限控制
	{
		systemResult.GET("/")        // 获取系统相关配置
		systemResult.PUT("/")        // 设置系统配置 key value
		systemResult.PUT("/title")   // 设置网站标题
		systemResult.PUT("/welcome") // 设置欢迎词
		systemResult.PUT("/footer")  // 设置网站脚注
		systemResult.PUT("/logo")    // 设置网站logo
	}

	// 帖子管理
	post := admin.Group("/post").Use(middleware.JWT().MiddlewareFunc()) // todo 权限控制
	{
		post.POST("/forum")            // 添加板块
		post.DELETE("/forum/:forumId") // 删除板块
		post.DELETE("/:postId")        // 删除帖子
	}

	// 用户管理
	user := admin.Group("/user").Use(middleware.JWT().MiddlewareFunc()) // todo 权限控制
	{
		user.PUT("/ban") // 禁用用户
	}
}
