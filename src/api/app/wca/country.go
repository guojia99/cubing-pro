package wca

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func Country(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		out := svc.Wca.CountryList()
		ctx.JSON(http.StatusOK, out)
	}
}
