package wca

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func PlayerResults(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		p := ctx.Param("wcaID")

		out, _ := svc.Wca.GetPersonResult(p)
		ctx.JSON(http.StatusOK, out)
	}
}

func PlayerCompetitions(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		p := ctx.Param("wcaID")

		out, _ := svc.Wca.GetPersonCompetition(p)
		ctx.JSON(http.StatusOK, out)
	}
}

func PlayerPersonInfo(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		p := ctx.Param("wcaID")

		out, err := svc.Wca.GetPersonInfo(p)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{})
			return
		}
		ctx.JSON(http.StatusOK, out)
	}
}
