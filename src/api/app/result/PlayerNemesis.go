package result

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type PlayerNemesisReq struct {
	CubeId string   `uri:"cubeId"`
	Events []string `json:"Events"`
}

func PlayerNemesis(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req PlayerNemesisReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
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

		nemesis := svc.Cov.PlayerNemesisWithID(usr.ID, req.Events)

		for i := 0; i < len(nemesis); i++ {
			nemesis[i].Single = nil
			nemesis[i].Avgs = nil
		}

		exception.ResponseOK(ctx, nemesis)
	}
}
