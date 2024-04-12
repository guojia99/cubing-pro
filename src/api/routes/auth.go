package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/guojia99/cubing-pro/src/api/app/auth"
	"github.com/guojia99/cubing-pro/src/svc"
)

func AuthRouters(router *gin.RouterGroup, svc *svc.Svc) {
	authG := router.Group("/auth") // 限流
	{
		authG.GET("/code", auth.Code(svc))           // 校验码
		authG.POST("/login", auth.Register(svc))     // 用户登录 / 获取权限token
		authG.POST("/logout")                        // 登出
		authG.POST("/register")                      // 用户注册
		authG.POST("/refresh")                       // 刷新秘钥
		authG.PUT("/user/:userId/retrieve/password") // 用户找回密码
	}
}
