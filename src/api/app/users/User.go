package users

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/public"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type UserBaseResultReq struct {
	PlayerId uint `uri:"playerId"`
}

func UserBaseResult(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var req UserBaseResultReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var usr user.User

		if err := svc.DB.First(&usr, "id = ?", req.PlayerId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, public.UserToUser(usr))
	}
}
