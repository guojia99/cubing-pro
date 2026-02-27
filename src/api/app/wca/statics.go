package wca

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/wca/types"
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

type GetEventRankWithTimerReq struct {
	EventID string `uri:"eventID"`

	Year    int    `json:"year"`
	Country string `json:"country"`
	IsAvg   bool   `json:"is_avg"`
	Page    int    `json:"page"`
	Size    int    `json:"size"`
}

type GetEventRankWithTimerResp struct {
	Data  []types.StaticWithTimerRank `json:"data"`
	Total int64                       `json:"total"`
}

func GetEventRankWithTimer(svc *svc.Svc) gin.HandlerFunc {

	cacheData := cache.New(5*time.Minute, 10*time.Minute)

	return func(ctx *gin.Context) {
		var request GetEventRankWithTimerReq
		if err := ctx.ShouldBindUri(&request); err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{})
			return
		}
		if err := ctx.ShouldBindJSON(&request); err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{})
			return
		}

		key := fmt.Sprintf("%s_%s_%+v_%d_%d_%d", request.EventID, request.Country, request.IsAvg, request.Page, request.Size, request.Year)
		getData, ok := cacheData.Get(key)
		if ok {
			ctx.JSON(http.StatusOK, getData)
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
			ctx.JSON(http.StatusNotFound, gin.H{})
			return
		}

		resp := GetEventRankWithTimerResp{
			Data:  out,
			Total: total,
		}
		cacheData.Set(key, resp, cache.DefaultExpiration)

		ctx.JSON(http.StatusOK, resp)
	}
}

type GetEventRankWithFullNowRequest struct {
	GetEventRankWithTimerReq
}

type GetEventRankWithFullNowResp struct {
	Data  []types.Result `json:"data"`
	Total int64          `json:"total"`
}

func GetEventRankWithFullNow(svc *svc.Svc) gin.HandlerFunc {
	cacheData := cache.New(5*time.Minute, 10*time.Minute)

	return func(ctx *gin.Context) {
		var request GetEventRankWithFullNowRequest
		if err := ctx.ShouldBindUri(&request); err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{})
			return
		}
		if err := ctx.ShouldBindJSON(&request); err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{})
			return
		}

		key := fmt.Sprintf("%s_%s_%+v_%d_%d_%d", request.EventID, request.Country, request.IsAvg, request.Page, request.Size, request.Year)
		getData, ok := cacheData.Get(key)
		if ok {
			ctx.JSON(http.StatusOK, getData)
			return
		}
		if len(request.EventID) > 7 || request.EventID == "" {
			ctx.JSON(http.StatusNotFound, gin.H{})
			return
		}
		
		out, total, err := svc.Wca.GetEventRankWithFullNow(
			request.EventID,
			request.Country,
			request.IsAvg,
			request.Page,
			request.Size,
		)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{})
			return
		}

		resp := GetEventRankWithFullNowResp{
			Data:  out,
			Total: total,
		}
		cacheData.Set(key, resp, cache.DefaultExpiration)

		ctx.JSON(http.StatusOK, resp)
	}
}
