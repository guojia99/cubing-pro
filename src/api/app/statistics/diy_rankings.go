package statistics

import (
	"fmt"
	"path"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/convenient/job"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

var diyRankingLock = sync.Mutex{}

func GetDiyRankingMaps(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var out []string
		if err := system.GetKeyJSONValue(svc.DB, job.DiyRankingsKey, &out); err != nil {
			exception.ErrUserNotFound.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, out)
	}
}

type AddDiyRankingMapReq struct {
	Key string `json:"name"`
}

// AddDiyRankingMap 添加一个 diy key
func AddDiyRankingMap(svc *svc.Svc) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		diyRankingLock.Lock()
		defer diyRankingLock.Unlock()

		var req AddDiyRankingMapReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		var in interface{}
		if err := system.GetKeyJSONValue(svc.DB, req.Key, &in); err == nil {
			exception.ErrResultBeUse.ResponseWithError(ctx, fmt.Errorf("资源已存在"))
			return
		}

		var diyRankings []string
		_ = system.GetKeyJSONValue(svc.DB, job.DiyRankingsKey, &diyRankings)
		diyRankings = append(diyRankings, req.Key)
		_ = system.SetKeyJSONValue(svc.DB, job.DiyRankingsKey, diyRankings, "")
		_ = system.SetKeyJSONValue(svc.DB, req.Key, []string{}, req.Key)

		exception.ResponseOK(ctx, nil)
	}
}

type UpdateDiyRankingMapPersonsReq struct {
	Key     string   `uri:"key" binding:"required"`
	Persons []string `json:"persons"`
}

// UpdateDiyRankingMapPersons 更新列表的名单
func UpdateDiyRankingMapPersons(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		diyRankingLock.Lock()
		defer diyRankingLock.Unlock()

		var req UpdateDiyRankingMapPersonsReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		var in []string
		if err := system.GetKeyJSONValue(svc.DB, req.Key, &in); err != nil {
			exception.ErrResultBeUse.ResponseWithError(ctx, fmt.Errorf("分组不存在"))
			return
		}
		_ = system.SetKeyJSONValue(svc.DB, req.Key, req.Persons, req.Key)
		exception.ResponseOK(ctx, nil)
	}
}

type DiyRankingsReq struct {
	Key string `uri:"key"`
}

func GetDiyRankingMapPersons(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req DiyRankingsReq

		if err := ctx.ShouldBindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		var out []string
		if err := system.GetKeyJSONValue(svc.DB, req.Key, &out); err != nil {
			exception.ErrUserNotFound.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, out)
	}
}

func DiyRankings(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req DiyRankingsReq
		if err := ctx.BindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		dataKey := path.Join(job.DiyRankingsKey, req.Key, "data")
		var data interface{}
		if err := system.GetKeyJSONValue(svc.DB, dataKey, &data); err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, data)
	}
}
