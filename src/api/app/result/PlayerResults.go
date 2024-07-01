package result

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

// 玩家所有成绩
type PlayerResultsReq struct {
	PlayerId uint `uri:"playerId"`
}

func PlayerResults(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req PlayerResultsReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var out []result.Results
		if err := svc.DB.Find(&out, "player_id = ? and ban = ?", req.PlayerId, false).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, out)
	}
}
