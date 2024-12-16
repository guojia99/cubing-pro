package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type EndCompReq struct {
	CompReq
}

func EndComp(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req EndCompReq
		if err := ctx.BindUri(&req); err != nil {
			return
		}
		org := ctx.Value(org_mid.OrgAuthMiddlewareKey).(user.Organizers)
		var comp competition.Competition
		if err := svc.DB.First(&comp, "id = ? and orgId = ?", req.CompId, org.ID).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		if comp.IsDone {
			exception.ResponseOK(ctx, nil)
			return
		}

		comp.IsDone = true
		svc.DB.Save(&comp)
		exception.ResponseOK(ctx, nil)
	}
}
