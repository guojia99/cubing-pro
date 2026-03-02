package wca

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/patrickmn/go-cache"
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

type BaseStaticsRequest struct {
	EventID string `uri:"eventID"`

	Year    int    `json:"year"`
	Country string `json:"country"`
	IsAvg   bool   `json:"is_avg"`
	Page    int    `json:"page"`
	Size    int    `json:"size"`

	MinAttempted int `json:"min_attempted"`
}

type BaseStaticsResponse struct {
	Data  interface{} `json:"data"`
	Total int64       `json:"total"`
}

func BaseStaticsWithEventAndCacheKey(svc *svc.Svc, funcKey string) gin.HandlerFunc {
	cacheData := cache.New(5*time.Minute, 10*time.Minute)
	return func(ctx *gin.Context) {
		var req BaseStaticsRequest
		if err := ctx.ShouldBindUri(&req); err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{})
			return
		}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{})
			return
		}
		key := fmt.Sprintf("%s_%s_%+v_%d_%d_%d_%d", req.EventID, req.Country, req.IsAvg, req.Page, req.Size, req.Year, req.MinAttempted)
		getData, ok := cacheData.Get(key)
		if ok {
			ctx.JSON(http.StatusOK, getData)
			return
		}

		var out interface{}
		var count int64
		var err error

		switch funcKey {
		case "GetEventRankWithTimer":
			out, count, err = svc.Wca.GetEventRankWithTimer(req.EventID, req.Country, req.Year, req.IsAvg, req.Page, req.Size)
		case "GetEventRankWithFullNow":
			out, count, err = svc.Wca.GetEventRankWithFullNow(req.EventID, req.Country, req.IsAvg, req.Page, req.Size)
		case "GetEventRankWithOnlyYear":
			out, count, err = svc.Wca.GetEventRankWithOnlyYear(req.EventID, req.Country, req.Year, req.IsAvg, req.Page, req.Size)
		case "GetEventSuccessRateResult":
			out, count, err = svc.Wca.GetEventSuccessRateResult(req.EventID, req.Country, req.MinAttempted, req.Page, req.Size)
		}
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{})
			return
		}

		resp := BaseStaticsResponse{
			Data:  out,
			Total: count,
		}
		cacheData.Set(key, resp, cache.DefaultExpiration)
		ctx.JSON(http.StatusOK, resp)
	}
}
