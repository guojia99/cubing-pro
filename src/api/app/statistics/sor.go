package statistics

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type SorReq struct {
	Page   int      `form:"page" json:"page" query:"page"`
	Size   int      `form:"size" json:"size" query:"size"`
	Events []string `json:"events"`
}

func Sor(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var req SorReq
		if err := ctx.ShouldBindQuery(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		if err := ctx.Bind(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		var events []event.Event
		if len(req.Events) == 0 {
			svc.DB.Find(&events, "is_wca = ?", true)
		} else {
			svc.DB.Find(&events, "is_wca = ? and id in ?", true, req.Events)
		}

		if req.Page <= 1 {
			req.Page = 1
		}
		if req.Size > 1000 {
			req.Size = 1000
		}
		if req.Size < 10 {
			req.Size = 10
		}

		result, total := svc.Cov.SelectKinChSor(req.Page, req.Size, events)
		exception.ResponseOK(ctx, app_utils.GenerallyListResp{
			Items: result,
			Total: int64(total),
		})
	}
}
