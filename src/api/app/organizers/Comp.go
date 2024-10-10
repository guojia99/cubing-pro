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

type CompReq struct {
	CompId uint `uri:"compId"`
}

func Comp(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CompReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var comp competition.Competition
		org := ctx.Value(org_mid.OrgAuthMiddlewareKey).(user.Organizers)
		if err := svc.DB.First(&comp, "id = ? and orgId = ?", req.CompId, org.ID).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, comp)
	}
}
