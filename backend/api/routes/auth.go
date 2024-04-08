package routes

import (
	"github.com/gin-gonic/gin"
)

func AuthRouters(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.GET("/code")                           // 校验码
		auth.POST("/login")                         // 用户登录 / 获取权限token
		auth.POST("/register")                      // 用户注册
		auth.POST("/refresh")                       // 刷新秘钥
		auth.PUT("/user/:userId/retrieve/password") // 用户找回密码
	}
}
