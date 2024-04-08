package routes

import "github.com/gin-gonic/gin"

func UserRouters(router *gin.RouterGroup) {
	user := router.Group("user")
	{
		user.GET("/")               // 详细信息
		user.GET("/auth_rule_list") // 规则权限列表
		user.POST("/detail")        // 修改用户信息
		user.POST("/password")      // 修改用户密码
		user.POST("/avatar")        // 修改用户头像
	}
}
