package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type GetCompResultReq struct {
	CompReq
}

func GetCompResult(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req GetCompResultReq
		if err := ctx.BindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		var results []result.Results

		if err := svc.DB.
			Where("comp_id = ?", req.CompId).
			Where("event_id = ?", ctx.Query("event_id")).
			Where("round_number = ?", ctx.Query("round_num")).
			Find(&results).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, "")
			return
		}

		result.SortResult(results)
		exception.ResponseOK(ctx, results)
	}
}
