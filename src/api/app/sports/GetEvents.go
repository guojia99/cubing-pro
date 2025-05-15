package sports

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/sports"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func ListSportEvents(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var events []sports.SportEvent
		if err := svc.DB.Find(&events).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, gin.H{"events": events})
	}
}
