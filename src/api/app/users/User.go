package users

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/public"
	_interface "github.com/guojia99/cubing-pro/src/internel/convenient/interface"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type UserBaseResultReq struct {
	CubeId string `uri:"cubeId"`
}

type UserBaseResultResp struct {
	public.User
	Detail _interface.UserResultDetail `json:"Detail"`
	Best   _interface.PlayerBestResult `json:"BestResults"`
}

func UserBaseResult(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req UserBaseResultReq
		if err := ctx.BindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		var usr user.User
		if err := svc.DB.First(&usr, "cube_id = ?", req.CubeId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		var rs []result.Results
		svc.DB.Where("user_id = ?", usr.ID).Where("ban = ?", false).Find(&rs)

		exception.ResponseOK(ctx, UserBaseResultResp{
			User:   public.UserToUser(usr),
			Detail: svc.Cov.SelectUserResultDetail(req.CubeId),
			Best:   svc.Cov.SelectBestResultsWithEventSortWithPlayer(usr.CubeID),
		})
	}
}
