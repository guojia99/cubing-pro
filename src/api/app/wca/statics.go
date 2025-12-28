package wca

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func GetPersonRankTimer(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		p := ctx.Param("wcaID")

		out, err := svc.Wca.GetPersonRankTimer(p)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"data": out,
		})
	}
}
