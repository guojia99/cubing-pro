package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/groupcomp"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"gorm.io/gorm"
	"time"
)

type AddCompResultReq struct {
	CompReq
	Results  []float64      `json:"Results"`
	CubeID   string         `json:"CubeID"`
	RoundNum int            `json:"Round"`
	EventID  string         `json:"EventId"`
	Penalty  result.Penalty `json:"Penalty"`
}

func AddCompResult(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req AddCompResultReq
		comp := ctx.Value(org_mid.CompMiddlewareKey).(competition.Competition)
		var bindErr error
		if comp.IsGroupCompetition() {
			bindErr = ctx.ShouldBindJSON(&req)
		} else {
			bindErr = app_utils.BindAll(ctx, &req)
		}
		if err := bindErr; err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		if !comp.IsGroupCompetition() {
			var u user.User
			if err := svc.DB.First(&u, "cube_id = ?", req.CubeID).Error; err != nil {
				exception.ErrResultCreate.ResponseWithError(ctx, err)
				return
			}
			if _, err := groupcomp.LegacyAddResult(svc.DB, comp, u, req.EventID, req.RoundNum, req.Results, req.Penalty); err != nil {
				exception.ErrResultCreate.ResponseWithError(ctx, err)
				return
			}
			exception.ResponseOK(ctx, nil)
			return
		}
		err := svc.DB.WithContext(ctx.Request.Context()).Transaction(func(tx *gorm.DB) error {
			c, err := groupcomp.Lock(tx, comp.ID)
			if err != nil {
				return err
			}
			var u user.User
			if err = tx.First(&u, "cube_id = ?", req.CubeID).Error; err != nil {
				return err
			}
			_, err = groupcomp.AddResult(tx, c, u, req.EventID, req.RoundNum, req.Results, req.Penalty, time.Now(), true)
			return err
		})
		if err != nil {
			exception.ErrResultCreate.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
