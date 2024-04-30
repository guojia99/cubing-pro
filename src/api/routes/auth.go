package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/svc"

	"github.com/guojia99/cubing-pro/src/api/app/auth"
	"github.com/guojia99/cubing-pro/src/api/middleware"
)

func AuthRouters(router *gin.RouterGroup, svc *svc.Svc) {
	authG := router.Group("/auth") // 限流
	{
		if svc.Cfg.GlobalConfig.Debug {
			authG.POST("/password_check", auth.PasswordCheck(svc)) // 调试用，密码生成工具
		}

		authG.GET("/code", middleware.Code().CodeRouter(svc))   // 校验码 todo 限流
		authG.POST("/login", middleware.JWT().LoginHandler)     // 用户登录 / 获取权限token
		authG.POST("/logout", middleware.JWT().LogoutHandler)   // 登出
		authG.POST("/register", auth.Register(svc))             // 用户注册
		authG.POST("/refresh", middleware.JWT().RefreshHandler) // 刷新秘钥

		authG.POST("/register/cube_id", auth.RegisterWithOldCubeID(svc)) // 用旧用户进行注册
		authG.PUT("/user/:userId/reset/password")                        // 用户重置密码
		authG.PUT("/user/:userId/retrieve/password")                     // 用户找回密码
	}

}
