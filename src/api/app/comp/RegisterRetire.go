package comp

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type RegisterRetireCompReq struct {
	CompReq
}

func RegisterRetire(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req RegisterRetireCompReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}

		var comp competition.Competition
		if err = svc.DB.First(&comp, "id = ?", req.CompId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		// 超出退赛时间，就不给退
		if comp.RegistrationCancelDeadlineTime != nil && time.Since(*comp.RegistrationCancelDeadlineTime) > 0 {
			exception.ErrRegisterField.ResponseWithError(ctx, "超出退赛时间范围")
			return
		}

		var reg competition.CompetitionRegistration
		if err = svc.DB.First(&reg, "comp_id = ? and user_id = ?", req.CompId, user.ID).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		// todo 退费
		reg.RetireTime = utils.PtrNow()
		if err = svc.DB.Save(&reg).Error; err != nil {
			exception.ErrResultUpdate.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, nil)

	}
}
