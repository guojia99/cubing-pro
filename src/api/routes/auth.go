package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"

	"github.com/guojia99/cubing-pro/src/api/app/auth"
	"github.com/guojia99/cubing-pro/src/api/middleware"
)

func AuthRouters(router *gin.RouterGroup, svc *svc.Svc) {
	authG := router.Group("/auth") // 限流
	{
		//if svc.Cfg.GlobalConfig.Debug {
		//	authG.POST("/password_check", auth.PasswordCheck(svc)) // 调试用，密码生成工具
		//}

		// 验证码
		authG.GET("/code", middleware.Code().CodeRouter()) // 校验码 todo 限流

		// 用户生命周期
		authG.POST("/login", middleware.JWT().LoginHandler)     // 用户登录 / 获取权限token
		authG.POST("/logout", middleware.JWT().LogoutHandler)   // 登出
		authG.POST("/refresh", middleware.JWT().RefreshHandler) // 刷新秘钥

		// 用户注册
		authG.POST(
			"/register/email_code",
			//middleware.Code().VerifyCodeMiddlewareFn(svc),
			auth.SendRegisterEmailCode(svc, user.RegisterWithEmail),
		) // email 验证 todo 限流
		authG.POST("/register", auth.Register(svc)) // 用户注册

		// 找回密码
		authG.POST("/retrieve/password/email_code", middleware.Code().VerifyCodeMiddlewareFn(svc), auth.RetrievePasswordSendCode(svc)) // email验证码 todo限流
		authG.POST("/retrieve/password/check_code", auth.CheckCode(svc))                                                               // 验证验证码有效性
		authG.POST("/retrieve/password", auth.RetrievePassword(svc))                                                                   // 用户找回密码

		// 用户操作
		authG.PUT("/reset/password", middleware.JWT().MiddlewareFunc(), middleware.Code().VerifyCodeMiddlewareFn(svc), auth.ResetPassword(svc)) // 用户重置密码
		authG.GET("/current", middleware.JWT().MiddlewareFunc(), middleware.CheckAuthMiddlewareFunc(user.AuthPlayer), auth.Current(svc))
	}

}
