package organizers

import (
	"fmt"

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
			exception.ErrResultDelete.ResponseWithError(ctx, "比赛已结束,无法删除此成绩")
			return
		}

		// 查看是否存在

		var res result.Results
		if err := svc.DB.Where("id = ?", req.ResultID).First(&res).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, "成绩不存在")
			return
		}

		var nextRes result.Results
		if err := svc.DB.
			Where("comp_id = ?", res.CompetitionID).
			Where("event_id = ?", res.EventID).
			Where("round_number =?", res.RoundNumber+1).
			Where("user_id = ?", res.UserID).
			First(&nextRes).Error; err == nil && nextRes.ID > 0 {
			exception.ErrResultUpdate.ResponseWithError(ctx, fmt.Errorf("%s 成绩存在，无法删除前一轮的成绩", nextRes.Round))
			return
		}

		// 查看下一轮的成绩是否存在
		if err := svc.DB.Model(&result.Results{}).Delete(&result.Results{}, "id = ?", req.ResultID).Error; err != nil {
			exception.ErrResultDelete.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
