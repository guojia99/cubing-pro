package result

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type PlayerRecordsReq struct {
	PlayerID uint `uri:"playerId"`
}

func PlayerRecords(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req PlayerRecordsReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		var records []result.Record
		svc.DB.Where("user_id = ?", req.PlayerID).Find(&records)
		exception.ResponseOK(ctx, records)
	}
}
