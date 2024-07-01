package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type ResetPasswordReq struct {
	Password  string `json:"password" binding:"required"`
	TimeStamp int64  `json:"timestamp" binding:"required"`
}

func ResetPassword(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}

		var req ResetPasswordReq
		if err = app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		// 验证密码是否有效
		password, err := utils.DePwdCode(req.Password, req.TimeStamp)
		if err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		// todo 本次的token需要做失效处理
		// 重置密码， 重置token
		user.Password = password
		svc.DB.Save(&user)
		middleware.JWT().RefreshHandler(ctx)
	}
}
