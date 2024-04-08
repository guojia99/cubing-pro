package routes

import "github.com/gin-gonic/gin"

func AdminRouters(router *gin.RouterGroup) {
	admin := router.Group("/admin")

	// 角色管理
	userRole := admin.Group("/user_role")
	{
		userRole.GET("/")                         // 角色列表
		userRole.POST("/")                        // 添加角色
		userRole.DELETE("/")                      // 删除角色
		userRole.PUT("/")                         // 修改角色
		userRole.GET("/:roleId/users")            // 角色对应用户列表
		userRole.PUT("/:roleId/bind")             // 用户绑定角色
		userRole.PUT("/:roleId/unbind")           // 用户解绑角色
		userRole.GET("/:roleId/rules")            // 角色权限列表
		userRole.POST("/:roleId/rules")           // 新增角色权限
		userRole.DELETE("/:roleId/rules/:ruleId") // 删除角色权限
		userRole.PUT("/:roleId/rules/:ruleId")    // 修改角色权限
	}

	// 项目管理
	event := admin.Group("/events")
	{
		event.GET("/")    // 项目列表
		event.POST("/")   // 新增项目
		event.DELETE("/") // 移除项目
		event.PUT("/")    // 修改项目
	}

	// 通知管理
	notify := admin.Group("/notify")
	{
		notify.GET("/")          // 通知列表
		notify.GET("/:notifyId") // 通知详情
		notify.POST("/")         // 添加通知
		notify.PUT("/")          // 修改通知
		notify.DELETE("/")       // 删除通知

		notify.PUT("/:notifyId/top")   // 置顶通知
		notify.PUT("/:notifyId/fixed") // 通知添加到侧边栏
	}

	// 系统配置
	systemResult := admin.Group("/system_result")
	{
		systemResult.GET("/")        // 获取系统相关配置
		systemResult.PUT("/")        // 设置系统配置 key value
		systemResult.PUT("/title")   // 设置网站标题
		systemResult.PUT("/welcome") // 设置欢迎词
		systemResult.PUT("/footer")  // 设置网站脚注
		systemResult.PUT("/logo")    // 设置网站logo
	}

	// 帖子管理
	post := admin.Group("/post")
	{
		post.POST("/forum")            // 添加板块
		post.DELETE("/forum/:forumId") // 删除板块
		post.DELETE("/:postId")        // 删除帖子
	}

	// 用户管理
	user := admin.Group("/user")
	{
		user.PUT("/ban") // 禁用用户
	}
}
