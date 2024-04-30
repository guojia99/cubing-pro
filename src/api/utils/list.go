package utils

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"gorm.io/gorm"
)

type (
	GenerallyListReq struct {
		Page   int               `form:"page"`
		Size   int               `form:"size"`
		Like   map[string]string `json:"like"`
		Search map[string]string `json:"search"`
	}
	GenerallyListResp struct {
		Items interface{} `json:"items"`
		Total int64       `json:"total"`
		Next  bool        `json:"next"`
	}
)

func GenerallyList(ctx *gin.Context, db *gorm.DB, model, dest interface{}, maxSize int, conds ...interface{}) {
	var req GenerallyListReq
	if err := ctx.Bind(&req); err != nil {
		exception.ErrRequestBinding.ResponseWithError(ctx, err)
		return
	}

	if req.Size > maxSize || req.Size <= 0 {
		req.Size = maxSize
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	// search with db
	searchDB := db.WithContext(ctx).Model(&model)
	if len(req.Like) > 0 {
		for key, val := range req.Like {
			searchDB = searchDB.Where(fmt.Sprintf("%s like ?", key), fmt.Sprintf("%%%s%%", val))
		}
	}
	if len(req.Search) > 0 {
		for key, val := range req.Search {
			searchDB = searchDB.Where(fmt.Sprintf("%s = ?", key), val)
		}
	}

	// total
	var total int64
	if err := searchDB.Count(&total).Error; err != nil {
		exception.ErrDatabase.ResponseWithError(ctx, err)
		return
	}

	// page
	offset := (req.Page - 1) * req.Size
	if err := searchDB.Offset(offset).Limit(req.Size).Find(&dest, conds...).Error; err != nil {
		exception.ErrDatabase.ResponseWithError(ctx, err)
		return
	}

	exception.ResponseOK(
		ctx, GenerallyListResp{
			Items: dest,
			Total: total,
		},
	)
}
