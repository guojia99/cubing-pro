package statistics

import (
	"path"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/convenient/job"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type DiyRankingsReq struct {
	Key string `uri:"key"`
}

func DiyRankings(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var req DiyRankingsReq
		if err := ctx.BindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		dataKey := path.Join(job.DiyRankingsKey, req.Key, "data")
		var data interface{}
		if err := system.GetKeyJSONValue(svc.DB, dataKey, &data); err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, data)
	}
}
