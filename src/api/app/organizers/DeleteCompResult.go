package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/groupcomp"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"strconv"
)

func DeleteCompResult(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseUint(ctx.Param("result_id"), 10, 32)
		if err != nil || id == 0 {
			exception.ErrRequestBinding.ResponseWithError(ctx, "无效成绩ID")
			return
		}
		c := ctx.Value(org_mid.CompMiddlewareKey).(competition.Competition)
		if err = groupcomp.DeleteResult(svc.DB.WithContext(ctx.Request.Context()), c.ID, uint(id)); err != nil {
			exception.ErrResultDelete.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
