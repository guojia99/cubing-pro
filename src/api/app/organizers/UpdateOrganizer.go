package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func UpdateOrganizers(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

type DoWithOrganizerReq struct {
	OrganizerReq
	Status       *user.OrganizersStatus `json:"Status,omitempty"`
	AdminMessage *string                `json:"AdminMessage"`
}

func DoWithOrganizers(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req DoWithOrganizerReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		var org user.Organizers
		if err := svc.DB.First(&org, "id = ?", req.OrgId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		if req.Status != nil {
			org.Status = *req.Status
		}
		if req.AdminMessage != nil {
			org.AdminMessage = *req.AdminMessage
		}

		svc.DB.Save(&org)
	}
}
