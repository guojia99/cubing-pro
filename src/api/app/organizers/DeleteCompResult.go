package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type DeleteCompResultReq struct {
	ResultID uint `uri:"result_id"`
}

func DeleteCompResult(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req DeleteCompResultReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		comp := ctx.Value(org_mid.CompMiddlewareKey).(competition.Competition)
		if !comp.IsRunningTime() { // todo 其他状态？
			exception.ErrResultDelete.ResponseWithError(ctx, "比赛已结束")
			return
		}

		if err := svc.DB.Model(&result.Results{}).Delete(&result.Results{}, "id = ?", req.ResultID).Error; err != nil {
			exception.ErrResultDelete.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
