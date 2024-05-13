package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type CheckCodeReq struct {
	UserID    string `json:"userID" binding:"required"`
	Key       string `json:"key" binding:"required"`
	Type      string `json:"type" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	EmailCode string `json:"emailCode" binding:"required"`
}

func CheckCode(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CheckCodeReq
		if err := ctx.ShouldBind(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		var checker user.CheckCode
		err := svc.DB.Where("use = ?", false).
			Where("key = ?", req.Key).
			Where("uid = ?", req.UserID).
			Where("email = ?", req.Email).
			Where("code = ?", req.EmailCode).
			Where("typ = ?", req.Type).First(&checker).Error

		if err != nil {
			exception.ErrAuthField.ResponseWithError(ctx, err)
			return
		}

		if time.Since(checker.Timeout) > 0 {
			exception.ErrVerifyCodeField.ResponseWithError(ctx, "验证码过期")
			return
		}
		exception.ResponseOK(ctx, "ok")
	}
}
