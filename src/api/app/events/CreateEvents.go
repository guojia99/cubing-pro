package events

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type CreateEventsReq struct {
	event.Event
}

func CreateEvents(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CreateEventsReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		if err := svc.DB.Create(&req.Event).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
