package events

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type DeleteEventReq struct {
	Id string `json:"id"`
}

func DeleteEvent(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req DeleteEventReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		var use int64
		svc.DB.Model(&result.Results{}).Where("event_id = ?", req.Id).Count(&use)

		if use > 0 {
			exception.ErrResultBeUse.ResponseWithError(ctx, fmt.Errorf("该项目目前含有 %d 个成绩，无法删除该项目", use))
			return
		}

		if err := svc.DB.Delete(&event.Event{}, "id = ?", req.Id).Error; err != nil {
			exception.ErrResultDelete.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
