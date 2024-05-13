package systemResults

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type setKeyValueReq struct {
	Key         string `uri:"key"`
	Value       string `json:"value" validate:"required"`
	Description string `json:"description" validate:"max=150"`
}

func setKeyValue(key string, svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req setKeyValueReq
		if err := ctx.ShouldBind(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		if key == "" {
			_ = ctx.BindUri(&req)
			if strings.Contains(req.Key, baseKey) {
				exception.ErrRequestBinding.ResponseWithError(ctx, fmt.Sprintf("不可使用带`%s`的kv", baseKey))
				return
			}
			key = req.Key
		}

		var kv = system.KeyValue{
			StringIDModel: basemodel.StringIDModel{
				ID: key,
			},
			Value:       req.Value,
			Description: req.Description,
		}

		if err := svc.DB.Save(&kv).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
