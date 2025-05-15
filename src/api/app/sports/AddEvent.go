package sports

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/sports"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func CreateSportEvent(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input sports.SportEvent
		if err := ctx.ShouldBindJSON(&input); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		if err := svc.DB.Create(&input).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, input)
	}
}
