package users

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type ResetPasswordReq struct {
	UserId uint `json:"userId"`
}

func RetrievePasswordWithAdmin(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var req ResetPasswordReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var userD user.User
		if err := svc.DB.First(&userD, "id = ?", req.UserId); err != nil {
			exception.ErrUserNotFound.ResponseWithError(ctx, err)
			return
		}

		checker := user.CheckCode{
			Type:    user.RetrievePasswordWithAdminKey,
			UserID:  userD.ID,
			Use:     false,
			Key:     utils.RandomString(32),
			Code:    "cubing-pro",
			Timeout: time.Now().Add(time.Hour * 6),
		}
		if err := svc.DB.Create(&checker).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(
			ctx, gin.H{
				"key": checker.Key,
			},
		)
	}
}
