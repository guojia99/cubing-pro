package app_utils

import (
	"fmt"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"gorm.io/gorm"
)

type (
	GenerallyListReq struct {
		Page      int               `form:"page"`
		Size      int               `form:"size"`
		Like      map[string]string `json:"like" query:"like"`
		Search    map[string]string `json:"search" query:"search"`
		StartTime int64             `form:"start_time" query:"start_time"`
		EndTime   int64             `form:"end_time" query:"end_time"`
		Order     []string          `json:"order"`
	}
	GenerallyListResp struct {
		Items interface{} `json:"items"`
		Total int64       `json:"total"`
		Next  bool        `json:"next"`
	}
)

type ListSearchParam struct {
	Model            interface{}
	MaxSize          int // 为0则代表全部
	Query            string
	QueryCons        []interface{}
	CanSearchAndLike []string // 允许查询的字段
	OrderBy          []string
	Omit             []string // 不需要字段
	Select           []string // 所需字段
	HasDeleted       bool     // 包含删除的行
	NotAutoResp      bool     //  自动封装消息
}

func GenerallyList(ctx *gin.Context, db *gorm.DB, dest interface{}, param ListSearchParam) {
	var req GenerallyListReq
	if err := ctx.Bind(&req); err != nil {
		exception.ErrRequestBinding.ResponseWithError(ctx, err)
		return
	}

	if req.Size > param.MaxSize || req.Size <= 0 {
		req.Size = param.MaxSize
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	// search with db
	//searchDB := db.WithContext(ctx).Model(&param.Model)
	searchDB := db.Model(&param.Model)

	if param.HasDeleted {
		searchDB = searchDB.Unscoped()
	}
	if len(req.Like) > 0 {
		for key, val := range req.Like {
			if slices.Contains(param.CanSearchAndLike, key) {
				searchDB = searchDB.Where(fmt.Sprintf("%s like ?", key), fmt.Sprintf("%%%s%%", val))
			}
		}
	}
	if len(req.Search) > 0 {
		for key, val := range req.Search {
			if slices.Contains(param.CanSearchAndLike, key) {
				searchDB = searchDB.Where(fmt.Sprintf("%s = ?", key), val)
			}
		}
	}
	if req.StartTime != 0 {
		searchDB = searchDB.Where("created_at >= ?", time.Unix(req.StartTime, 0))
	}
	if req.EndTime != 0 {
		searchDB = searchDB.Where("created_at <= ?", time.Unix(req.EndTime, 0))
	}

	if len(param.Query) != 0 {
		searchDB = searchDB.Where(param.Query, param.QueryCons...)
	}

	// order by 默认倒序
	if len(param.OrderBy) > 0 {
		for _, o := range param.OrderBy {
			searchDB = searchDB.Order(o)
		}
	}
	if len(req.Order) > 0 {
		for _, o := range req.Order {
			searchDB = searchDB.Order(o)
		}
	} else {
		searchDB = searchDB.Order("created_at DESC")
	}

	// total
	var total int64
	if err := searchDB.Count(&total).Error; err != nil {
		exception.ErrDatabase.ResponseWithError(ctx, err)
		return
	}

	// omit
	if len(param.Omit) > 0 {
		searchDB = searchDB.Omit(param.Omit...)
	}
	if len(param.Select) > 0 {
		param.Select = append(param.Select, "id", "created_at", "updated_at", "deleted_at")
		searchDB = searchDB.Select(param.Select)
	}

	// page
	offset := (req.Page - 1) * req.Size
	var err error
	if param.MaxSize == 0 {
		err = searchDB.Find(&dest).Error
	} else {
		err = searchDB.Offset(offset).Limit(req.Size).Find(&dest).Error
	}
	if err != nil {
		exception.ErrDatabase.ResponseWithError(ctx, err)
		return
	}

	if !param.NotAutoResp {
		exception.ResponseOK(
			ctx, GenerallyListResp{
				Items: dest,
				Total: total,
			},
		)
	}
}
