package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/groupcomp"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"time"
)

type RefreshEventReq struct {
	All         bool    `json:"All"`
	EventID     *string `json:"EventID"`
	RoundNumber *int    `json:"RoundNumber"`
	Open        *bool   `json:"Open"`
}

func RefreshEvent(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req RefreshEventReq
		c := ctx.Value(org_mid.CompMiddlewareKey).(competition.Competition)
		var bindErr error
		if c.IsGroupCompetition() {
			bindErr = ctx.ShouldBindJSON(&req)
		} else {
			bindErr = app_utils.BindAll(ctx, &req)
			if req.Open == nil {
				v := false
				req.Open = &v
			}
		}
		if bindErr != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, bindErr)
			return
		}
		if req.All {
			if !c.IsGroupCompetition() || req.Open == nil || !*req.Open || req.EventID != nil || req.RoundNumber != nil {
				exception.ErrRequestBinding.ResponseWithError(ctx, "All仅支持群赛，必须Open=true且不能提供EventID/RoundNumber")
				return
			}
			if err := groupcomp.OpenAll(svc.DB.WithContext(ctx.Request.Context()), c.ID, time.Now()); err != nil {
				exception.ErrResultUpdate.ResponseWithError(ctx, err)
				return
			}
			exception.ResponseOK(ctx, nil)
			return
		}
		if req.Open == nil || (c.IsGroupCompetition() && (req.EventID == nil || *req.EventID == "" || req.RoundNumber == nil || *req.RoundNumber < 1)) {
			exception.ErrRequestBinding.ResponseWithError(ctx, "必须提供EventID/RoundNumber/Open")
			return
		}
		eventID, round := "", 0
		if req.EventID != nil {
			eventID = *req.EventID
		}
		if req.RoundNumber != nil {
			round = *req.RoundNumber
		}
		if err := groupcomp.Refresh(svc.DB.WithContext(ctx.Request.Context()), c.ID, eventID, round, *req.Open, time.Now()); err != nil {
			exception.ErrResultUpdate.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
