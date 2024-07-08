package comp

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func Results(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CompReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		var results []result.Results
		svc.DB.Where("comp_id = ?", req.CompId).Find(&results)
		exception.ResponseOK(ctx, results)
	}
}
