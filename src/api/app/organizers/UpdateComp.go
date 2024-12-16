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

type UpdateCompReq struct {
	CompReq
	competition.Competition
}

func UpdateComp(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req UpdateCompReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		org := ctx.Value(org_mid.OrgAuthMiddlewareKey).(user.Organizers)
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
		comp.CompJSON = req.CompJSON
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

type ApprovalCompReq struct {
	CompReq
	Ok bool `json:"Ok"`
}

func ApprovalComp(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req ApprovalCompReq
		if err := ctx.BindUri(&req); err != nil {
			return
		}

		var comp competition.Competition

		if err := svc.DB.First(&comp, "id = ?", req.CompId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		req.Ok = true

		switch comp.Status {
		case competition.Reviewing:
			if req.Ok {
				comp.Status = competition.Running
			} else {
				comp.Status = competition.Reject
			}

			if err := svc.DB.Save(&comp).Error; err != nil {
				exception.ErrResultUpdate.ResponseWithError(ctx, err)
				return
			}
			// todo 发邮箱
		}
		exception.ResponseOK(ctx, nil)
	}
}
