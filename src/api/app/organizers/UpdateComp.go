package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type UpdateCompReq struct {
	CompReq
	competition.Competition
}

func UpdateComp(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req UpdateCompReq
		if err := ctx.BindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		org := ctx.Value(OrgAuthMiddlewareKey).(user.Organizers)
		var comp competition.Competition
		if err := svc.DB.First(&comp, "id = ? and orgId = ?", req.CompId, org.ID).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		comp.Illustrate = req.Illustrate
		comp.Location = req.Location
		comp.LocationAddr = req.LocationAddr
		comp.Country = req.Country
		comp.City = req.City
		comp.RuleMD = req.RuleMD
		comp.Events = req.Events
		comp.AutomaticReview = req.AutomaticReview
		comp.WCAUrl = req.WCAUrl

		comp.MinCount = req.MinCount
		comp.Count = req.Count
		comp.FreeParticipate = req.FreeParticipate
		comp.AutomaticReview = req.AutomaticReview

		comp.CompStartTime = req.CompStartTime
		comp.CompEndTime = req.CompEndTime
		comp.RegistrationRestartTime = req.RegistrationRestartTime
		comp.RegistrationEndTime = req.RegistrationEndTime
		comp.RegistrationCancelDeadlineTime = req.RegistrationCancelDeadlineTime
		comp.RegistrationRestartTime = req.RegistrationRestartTime

		if err := svc.DB.Save(&comp).Error; err != nil {
			exception.ErrResultUpdate.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, comp)
	}
}
