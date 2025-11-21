package wca

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/convenient/job"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func ChinaComps(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var out interface{}
		err := system.GetKeyJSONValue(svc.DB, job.UpdateCubingCompetitionKey, &out)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{})
			return
		}
		exception.ResponseOK(ctx, out)
	}
}
