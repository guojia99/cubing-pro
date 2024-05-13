package users

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	user2 "github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type BanUserReq struct {
	UserId    uint   `json:"UserIds"`
	Ban       bool   `json:"Ban"`
	BanReason string `json:"BanReason"`
}

func BanUser(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req BanUserReq
		if err := ctx.ShouldBind(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		var user user2.User
		if err := svc.DB.First(&user, "id = ?", req.UserId).Error; err != nil {
			exception.ErrUserNotFound.ResponseWithError(ctx, err)
			return
		}

		user.Ban = req.Ban
		user.BanReason = req.BanReason

		// todo 清理cache
		if err := svc.DB.Save(&user); err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
