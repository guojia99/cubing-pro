package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/groupcomp"
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
		req.CompId = ctx.Value(org_mid.CompMiddlewareKey).(competition.Competition).ID
		org := ctx.Value(org_mid.OrgAuthMiddlewareKey).(user.Organizers)
		if err := svc.DB.First(&comp, "id = ? and orgId = ?", req.CompId, org.ID).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		if comp.IsGroupCompetition() {
			orgName, groupName, err := groupcomp.CompetitionNames(svc.DB.WithContext(ctx.Request.Context()), comp.OrganizersID, comp.GroupID)
			if err != nil {
				exception.ErrDatabase.ResponseWithError(ctx, err)
				return
			}
			exception.ResponseOK(ctx, struct {
				competition.Competition
				OrganizersName string `json:"OrganizersName"`
				GroupName      string `json:"GroupName"`
			}{comp, orgName, groupName})
			return
		}
		exception.ResponseOK(ctx, comp)
	}
}
