package result

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type DeletePreResultReq struct {
	PreId uint `uri:"pre_id"`
}

func DeletePreResult(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req DeletePreResultReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		var pre result.PreResults
		if err := svc.DB.First(&pre, "id = ?", req.PreId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		if pre.Finish {
			exception.ErrResultDelete.ResponseWithError(ctx, "该预录入成绩已经审核完成，无法删除")
			return
		}
		if err := svc.DB.Delete(&pre, "id = ?", pre.ID).Error; err != nil {
			exception.ErrResultDelete.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, nil)
	}
}
