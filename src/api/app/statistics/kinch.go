package statistics

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type KinChReq struct {
	Page int `form:"page" json:"page" query:"page"`
	Size int `form:"size" json:"size" query:"size"`

	Age    int      `form:"age" query:"age"`
	Events []string `json:"events" query:"events"`
}

func KinCh(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var req KinChReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var events []event.Event
		if len(req.Events) == 0 {
			svc.DB.Find(&events, "is_wca = ?", true)
		} else {
			svc.DB.Find(&events, "is_wca = ? and id in ?", true, req.Events)
		}

		result, total := svc.Cov.SelectKinChSor(req.Page, req.Size, events)
		exception.ResponseOK(ctx, app_utils.GenerallyListResp{
			Items: result,
			Total: int64(total),
		})
	}
}

func SeniorKinCh(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req KinChReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		var events []event.Event
		if len(req.Events) == 0 {
			svc.DB.Find(&events, "is_wca = ?", true)
		} else {
			svc.DB.Find(&events, "is_wca = ? and id in ?", true, req.Events)
		}

		result, total := svc.Cov.SelectSeniorKinChSor(req.Page, req.Size, req.Age, events)
		exception.ResponseOK(ctx, app_utils.GenerallyListResp{
			Items: result,
			Total: int64(total),
		})
	}
}

type DiyRankingsKinchReq struct {
	DiyRankingsReq
	KinChReq
}

func DiyRankingsKinch(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req DiyRankingsKinchReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var events []event.Event
		if len(req.Events) == 0 {
			svc.DB.Find(&events, "is_wca = ?", true)
		} else {
			svc.DB.Find(&events, "is_wca = ? and id in ?", true, req.Events)
		}

		var wcaIds []string
		if err := system.GetKeyJSONValue(svc.DB, req.Key, &wcaIds); err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		result, total := svc.Cov.SelectKinchWithWcaIDs(wcaIds, req.Page, req.Size, events)
		exception.ResponseOK(ctx, app_utils.GenerallyListResp{
			Items: result,
			Total: int64(total),
		})
	}
}
