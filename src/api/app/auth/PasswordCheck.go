package auth

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type PasswordCheckRequest struct {
	Password  string `json:"password"`  // 密码（加密前）
	TimeStamp int64  `json:"timestamp"` // 创建时间戳
}

func PasswordCheck(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req PasswordCheckRequest
		if err := ctx.Bind(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		key := utils.GenerateRandomKey(req.TimeStamp)
		fmt.Println(key, req.TimeStamp)
		password, err := utils.Encrypt(req.Password, key)
		if err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, gin.H{"password": password})
	}
}
