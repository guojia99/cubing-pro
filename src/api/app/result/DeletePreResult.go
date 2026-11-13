package result

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	"github.com/guojia99/cubing-pro/src/internel/groupcomp"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"strconv"
)

func DeletePreResult(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseUint(ctx.Param("pre_id"), 10, 32)
		if err != nil || id == 0 {
			exception.ErrRequestBinding.ResponseWithError(ctx, "无效预成绩ID")
			return
		}
		u, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}
		if err = groupcomp.DeletePending(svc.DB.WithContext(ctx.Request.Context()), uint(id), u.ID); err != nil {
			exception.ErrResultDelete.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
