package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/groupcomp"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func CompetitionFilters(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req struct {
			Page int `form:"page"`
			Size int `form:"size"`
		}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		out, err := groupcomp.FilterOptions(svc.DB.WithContext(ctx.Request.Context()), req.Page, req.Size)
		if err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, out)
	}
}
