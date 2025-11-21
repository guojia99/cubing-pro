package statistics

import (
	"fmt"
	"path"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	_interface "github.com/guojia99/cubing-pro/src/internel/convenient/interface"
	"github.com/guojia99/cubing-pro/src/internel/convenient/job"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

var diyRankingLock = sync.Mutex{}

func GetDiyRankingMaps(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var keys []string
		if err := system.GetKeyJSONValue(svc.DB, job.DiyRankingsKey, &keys); err != nil {
			exception.ErrUserNotFound.ResponseWithError(ctx, err)
			return
		}

		var out []system.KeyValue
		svc.DB.Where("id in ?", keys).Find(&out)
		exception.ResponseOK(ctx, out)
	}
}

type AddDiyRankingMapReq struct {
	Key         string `json:"key"`
	Description string `json:"description"`
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
		_ = system.SetKeyJSONValue(svc.DB, req.Key, []string{}, req.Description)
		exception.ResponseOK(ctx, nil)
	}
}

type UpdateDiyRankingMapPersonsReq struct {
	Key         string   `uri:"key" binding:"required"`
	Persons     []string `json:"persons"`
	Description string   `json:"description"`
}

// UpdateDiyRankingMapPersons 更新列表的名单
func UpdateDiyRankingMapPersons(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		diyRankingLock.Lock()
		defer diyRankingLock.Unlock()

		var req UpdateDiyRankingMapPersonsReq
		if err := ctx.ShouldBindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		if err := ctx.Bind(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		var in []string
		if err := system.GetKeyJSONValue(svc.DB, req.Key, &in); err != nil {
			exception.ErrResultBeUse.ResponseWithError(ctx, fmt.Errorf("分组不存在"))
			return
		}
		err := system.SetKeyJSONValue(svc.DB, req.Key, req.Persons, req.Description)
		if err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
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

type DiyRankingSorReq struct {
	DiyRankingsReq
	KinChReq

	WithSingle bool `form:"withSingle" json:"withSingle" query:"withSingle"`
	WithAvg    bool `form:"withAvg" json:"withAvg" query:"withAvg"`
}

func DiyRankingSor(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req DiyRankingSorReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		opt := _interface.SelectSorWithWcaIDsOption{
			Events:     req.Events,
			WithSingle: req.WithSingle,
			WithAvg:    req.WithAvg,
		}
		if !opt.WithSingle && !opt.WithAvg {
			opt.WithAvg = true
		}
		if len(opt.Events) == 0 {
			var allWcaEvents []event.Event
			svc.DB.Where("is_wca = ?", true).Find(&allWcaEvents)
			for _, ev := range allWcaEvents {
				opt.Events = append(opt.Events, ev.ID)
			}
		}

		var wcaIds []string
		if err := system.GetKeyJSONValue(svc.DB, req.Key, &wcaIds); err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		result, total := svc.Cov.SelectSorWithWcaIDs(wcaIds, req.Page, req.Size, opt)
		exception.ResponseOK(ctx, app_utils.GenerallyListResp{
			Items: result,
			Total: int64(total),
		})
	}
}
