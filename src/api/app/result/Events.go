package result

import (
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type EventsResp struct {
	Events     []event.Event `json:"Events"`
	UpdateTime time.Time     `json:"UpdateTime"`
}

func Events(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var out []event.Event

		if err := svc.DB.Find(&out).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		sort.Slice(
			out, func(i, j int) bool {
				return out[i].Idx < out[j].Idx
			},
		)

		exception.ResponseOK(
			ctx, EventsResp{
				Events:     out,
				UpdateTime: time.Now(),
			},
		)

	}
}
