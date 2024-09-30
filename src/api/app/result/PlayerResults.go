package result

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

// 玩家所有成绩
type PlayerResultsReq struct {
	CubeId string `uri:"cubeId"`
}

type PlayerResultsResp struct {
	All []result.Results
}

func PlayerResults(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req PlayerResultsReq
		if err := ctx.BindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		//if err := app_utils.BindAll(ctx, &req); err != nil {
		//	return
		//}
		var usr user.User
		if err := svc.DB.First(&usr, "cube_id = ?", req.CubeId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		var rs []result.Results
		if err := svc.DB.Find(&rs, "cube_id = ? and ban = ?", req.CubeId, false).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, PlayerResultsResp{
			All: rs,
		})
	}
}
