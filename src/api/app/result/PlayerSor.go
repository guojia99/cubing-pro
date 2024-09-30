package result

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type PlayerSorReq struct {
	CubeId string   `uri:"cubeId"`
	Events []string `json:"Events"`
}

func PlayerSor(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req PlayerSorReq
		if err := ctx.BindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		if err := ctx.Bind(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		var usr user.User
		if err := svc.DB.First(&usr, "cube_id = ?", req.CubeId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		if len(req.Events) == 0 {
			var eventIds []string
			svc.DB.Model(&event.Event{}).Distinct("id").Where("is_wca = ?", true).Find(&eventIds)
			req.Events = eventIds
		}

		sor, err := svc.Cov.KinChSorWithPlayer(usr.ID, req.Events)
		if err != nil {
			exception.ErrGetData.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, sor)
	}
}
