package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func DeleteComp(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CompReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		org := ctx.Value(org_mid.OrgAuthMiddlewareKey).(user.Organizers)
		var comp competition.Competition
		if err := svc.DB.First(&comp, "id = ? and orgId = ?", req.CompId, org.ID).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		switch comp.Status {
		case competition.Reviewing, competition.Temporary, competition.Reject:
			svc.DB.Delete(&comp)
		}

		exception.ResponseOK(ctx, comp)
	}
}
