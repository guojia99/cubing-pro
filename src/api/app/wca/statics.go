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

type GetEventRankWithTimerReq struct {
	EventID string `uri:"eventID"`

	Year    int    `form:"year"`
	Country string `form:"country"`
	IsAvg   bool   `form:"is_avg"`
	Page    int    `form:"page"`
	Size    int    `form:"size"`
}

func GetEventRankWithTimer(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var request GetEventRankWithTimerReq
		if err := ctx.ShouldBindUri(&request); err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{})
			return
		}

		out, total, err := svc.Wca.GetEventRankWithTimer(
			request.EventID,
			request.Country,
			request.Year,
			request.IsAvg,
			request.Page,
			request.Size,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"data":  out,
			"total": total,
		})
	}
}
