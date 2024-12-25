package result

import (
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type PlayerRecordsReq struct {
	CubeId string `uri:"cubeId"`
}

func PlayerRecords(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req PlayerRecordsReq
		if err := ctx.BindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		var records []result.Record
		svc.DB.Where("cube_id = ?", req.CubeId).Find(&records)

		sort.Slice(records, func(i, j int) bool {
			return records[i].CompsId > records[j].CompsId
		})
		exception.ResponseOK(ctx, records)
	}
}
