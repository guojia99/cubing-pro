package comp

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func Results(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CompReq
		if err := ctx.BindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		fmt.Println(req.CompId)
		var results []result.Results
		svc.DB.Where("comp_id = ?", req.CompId).Find(&results)
		exception.ResponseOK(ctx, results)
	}
}
