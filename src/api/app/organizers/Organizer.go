package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type OrganizerReq struct {
	OrgId uint `uri:"orgId"`
}

func Organizer(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		org := ctx.Value(OrgAuthMiddlewareKey).(user.Organizers)
		exception.ResponseOK(ctx, org)
	}
}
