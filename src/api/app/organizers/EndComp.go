package organizers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/groupcomp"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func EndComp(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c := ctx.Value(org_mid.CompMiddlewareKey).(competition.Competition)
		if err := groupcomp.End(svc.DB.WithContext(ctx.Request.Context()), c.ID, time.Now); err != nil {
			exception.ErrResultUpdate.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
