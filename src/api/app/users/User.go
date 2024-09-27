package users

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/public"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type UserBaseResultReq struct {
	PlayerId string `uri:"playerId"`
}

type UserBaseDetail struct {
	RestoresNum  int `json:"RestoresNum"`
	SuccessesNum int `json:"SuccessesNum"`
	Matches      int `json:"Matches"`
	PodiumNum    int `json:"PodiumNum"`
}

type UserBaseResults struct {
}

type UserBaseResultResp struct {
	public.User
	Detail UserBaseDetail  `json:"Detail"`
	Result UserBaseResults `json:"Result"`
}

func UserBaseResult(svc *svc.Svc) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var req UserBaseResultReq
		if err := ctx.BindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		var usr user.User
		if err := svc.DB.First(&usr, "cube_id = ?", req.PlayerId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, UserBaseResultResp{
			User: public.UserToUser(usr),
		})
	}
}
