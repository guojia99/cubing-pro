package events

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func Events(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var events []event.Event
		svc.DB.Find(&events)

		exception.ResponseOK(ctx, events)
	}
}
