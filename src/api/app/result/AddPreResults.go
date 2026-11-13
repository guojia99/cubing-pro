package result

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/groupcomp"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"time"
)

type AddPreResultsReq struct {
	CompId  uint           `json:"CompId"`
	Round   string         `json:"Round"`
	Results []float64      `json:"Results"`
	Penalty result.Penalty `json:"Penalty"`
	EventID string         `json:"EventID"`
}

func AddPreResults(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req AddPreResultsReq
		if err := ctx.ShouldBind(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		u, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}
		var c competition.Competition
		if err = svc.DB.First(&c, "id = ?", req.CompId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		if !c.IsGroupCompetition() {
			_, err = groupcomp.LegacySavePending(svc.DB, c, u, req.EventID, req.Round, req.Results, req.Penalty)
			if err != nil {
				exception.ErrResultCreate.ResponseWithError(ctx, err)
				return
			}
			exception.ResponseOK(ctx, nil)
			return
		}
		ev, s, err := groupcomp.Validate(c, req.EventID, req.Round, u.ID, time.Now(), true, true)
		if err != nil {
			exception.ErrResultCreate.ResponseWithError(ctx, err)
			return
		}
		pres := []result.PreResults{{Results: result.Results{CompetitionID: c.ID, CompetitionName: c.Name, Round: s.Round, RoundNumber: s.RoundNum, PersonName: u.Name, UserID: u.ID, CubeID: u.CubeID, Result: req.Results, Penalty: req.Penalty, EventID: ev.EventID, EventName: ev.EventName, EventRoute: ev.EventRoute}, CompsName: c.Name, RoundName: s.Round, Recorder: u.Name, Source: "web-api"}}
		if err = groupcomp.SavePending(svc.DB.WithContext(ctx.Request.Context()), pres); err != nil {
			exception.ErrResultCreate.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, pres[0])
	}
}
