package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type ApprovalCompPlayerPreResultReq struct {
	ResultID uint `uri:"result_id"`

	Detail string `json:"FinishDetail"`
}

func ApprovalCompPlayerPreResult(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req ApprovalCompPlayerPreResultReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var pre result.PreResults
		if err := svc.DB.First(&pre, "id = ?", req.ResultID).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}

		//org := ctx.Value(or)
		comp := ctx.Value(org_mid.CompMiddlewareKey).(competition.Competition)
		if !comp.IsRunningTime() {
			exception.ErrResultCreate.ResponseWithError(ctx, "不在比赛时间")
			return
		}
		if pre.Finish {
			exception.ErrResultCreate.ResponseWithError(ctx, "成绩已被处理，请不要反复处理")
			return
		}

		pre.Processor = user.Name
		pre.ProcessorID = user.ID
		pre.Detail = req.Detail
		if req.Detail == result.DetailOk || req.Detail == result.DetailNot {
			pre.FinishDetail = req.Detail
			pre.Finish = true
		}

		if req.Detail == result.DetailOk {
			// 判断选手是否已经审核完成
			res, err := checkAndAddPlayerResult(
				ctx, svc, AddCompResultReq{
					CompReq:  CompReq{CompId: comp.ID},
					Results:  pre.Result,
					CubeID:   pre.CubeID,
					RoundNum: pre.RoundNumber,
					EventID:  pre.EventID,
					Penalty:  pre.Penalty,
				},
			)
			if err != nil {
				exception.ErrResultCreate.ResponseWithError(ctx, err)
				return
			}
			pre.ResultID = &res.ID
		}
		svc.DB.Save(&pre)
		exception.ResponseOK(ctx, nil)
	}
}
