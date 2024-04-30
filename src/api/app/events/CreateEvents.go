package events

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type CreateEventsReq struct {
	event.Event
}

func CreateEvents(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CreateEventsReq
		if err := ctx.Bind(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		svc.DB.Create(req.Event)
		exception.ResponseOK(ctx, nil)
	}
}
