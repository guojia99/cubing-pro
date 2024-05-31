package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type CompPlayerApprovalReq struct {
	CompReq

	RegId uint `uri:"reg_id"`
	Pass  bool `json:"pass"`
}

func CompPlayerApproval(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CompPlayerApprovalReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		var reg competition.CompetitionRegistration
		if err := svc.DB.First(&reg, "id = ?", req.RegId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		reg.Status = competition.RegisterStatusNotApply
		if req.Pass {
			reg.Status = competition.RegisterStatusPass
		}
		if err := svc.DB.Save(&reg).Error; err != nil {
			exception.ErrResultUpdate.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
