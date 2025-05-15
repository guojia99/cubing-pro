package sports

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/sports"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func DeleteSportEvent(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		if err := svc.DB.Delete(&sports.SportEvent{}, "id = ?", id).Error; err != nil {
			exception.ErrResultDelete.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, gin.H{"id": id})
	}
}
