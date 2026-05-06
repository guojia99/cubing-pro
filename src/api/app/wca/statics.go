package wca

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
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
	EventID string `uri:"eventID" json:"eventID"`

	Year         int      `json:"year"`
	Month        int      `json:"month"`
	Country      string   `json:"country"`
	IsAvg        bool     `json:"is_avg"`
	Page         int      `json:"page"`
	Size         int      `json:"size"`
	Events       []string `json:"events"`
	MinAttempted int      `json:"min_attempted"`

	LackNum int `json:"lackNum"`
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

		key, err := utils.MakeCacheKey(req)
		if err != nil {
			exception.ErrGetData.ResponseWithError(ctx, err)
			return
		}
		getData, ok := cacheData.Get(key)
		if ok {
			ctx.JSON(http.StatusOK, getData)
			return
		}

		var out interface{}
		var count int64

		fmt.Println(funcKey)
		switch funcKey {
		case "GetEventRankWithTimer":
			out, count, err = svc.Wca.GetEventRankWithTimer(req.EventID, req.Country, req.Year, req.IsAvg, req.Page, req.Size)
		case "GetEventRankWithFullNow":
			out, count, err = svc.Wca.GetEventRankWithFullNow(req.EventID, req.Country, req.IsAvg, req.Page, req.Size)
		case "GetEventRankWithOnlyYear":
			out, count, err = svc.Wca.GetEventRankWithOnlyYear(req.EventID, req.Country, req.Year, req.Month, req.IsAvg, req.Page, req.Size)
		case "GetEventSuccessRateResult":
			out, count, err = svc.Wca.GetEventSuccessRateResult(req.EventID, req.Country, req.MinAttempted, req.Page, req.Size)
		case "GetAllEventsAchievement":
			out, count, err = svc.Wca.GetAllEventsAchievement(req.LackNum, req.Country, req.Page, req.Size)
		case "GetRankWithEvents":
			out, count, err = svc.Wca.GetRankWithEvents(req.Events, req.Country, req.IsAvg, req.Page, req.Size)
		case "GetWithCompYearPersonRank":
			out, count, err = svc.Wca.GetWithCompYearPersonRank(req.Year, req.Country, req.EventID, req.IsAvg, req.Page, req.Size)

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

func GetGrandSlam(svc *svc.Svc) gin.HandlerFunc {
	cacheData := cache.New(5*time.Minute, 10*time.Minute)
	return func(ctx *gin.Context) {
		if val, ok := cacheData.Get("grandslam"); ok {
			ctx.JSON(http.StatusOK, val)
			return
		}
		data := svc.Wca.GetGrandSlam()
		ctx.JSON(http.StatusOK, data)
		cacheData.Set("grandslam", data, cache.DefaultExpiration)
	}
}
