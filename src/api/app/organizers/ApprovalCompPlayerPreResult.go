package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/groupcomp"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"strconv"
)

type ApprovalCompPlayerPreResultReq struct {
	Detail              string `json:"FinishDetail"`
	HistoricalConfirmed bool   `json:"HistoricalConfirmed"`
	ReviewReason        string `json:"ReviewReason"`
}

func ApprovalCompPlayerPreResult(svc *svc.Svc) gin.HandlerFunc {
	return approvalCompPlayerPreResult(svc, false)
}

func AdminApprovalCompPlayerPreResult(svc *svc.Svc) gin.HandlerFunc {
	return approvalCompPlayerPreResult(svc, true)
}

func approvalCompPlayerPreResult(svc *svc.Svc, adminRoute bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req ApprovalCompPlayerPreResultReq
		c := ctx.Value(org_mid.CompMiddlewareKey).(competition.Competition)
		var bindErr error
		if c.IsGroupCompetition() {
			bindErr = ctx.ShouldBindJSON(&req)
		} else {
			bindErr = app_utils.BindAll(ctx, &req)
		}
		if err := bindErr; err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		if req.HistoricalConfirmed && !adminRoute {
			exception.ErrAuthField.ResponseWithError(ctx, "历史人工核查确认只能通过管理员群赛审批接口执行")
			return
		}
		id, err := strconv.ParseUint(ctx.Param("result_id"), 10, 32)
		if err != nil || id == 0 {
			exception.ErrRequestBinding.ResponseWithError(ctx, "无效成绩ID")
			return
		}
		u, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}
		if adminRoute {
			err = groupcomp.AdminApprove(svc.DB.WithContext(ctx.Request.Context()), c.ID, uint(id), req.Detail, u, req.HistoricalConfirmed, req.ReviewReason)
		} else {
			err = groupcomp.Approve(svc.DB.WithContext(ctx.Request.Context()), c.ID, uint(id), req.Detail, u)
		}
		if err != nil {
			exception.ErrResultUpdate.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
