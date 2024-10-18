package statics

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"net/http"
	"path"
)

type ImageReq struct {
	Uid string `uri:"uid" json:"uid"`
}

func Image(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var req ImageReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var img system.Image
		if svc.DB.First(&img, "uid = ?", req.Uid).Error != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, "无数据")
			return
		}

		if len(img.URL) != 0 {
			ctx.Redirect(http.StatusMovedPermanently, img.URL)
			return
		}

		if len(img.LocalPath) != 0 {
			ctx.Header("Content-Type", fmt.Sprintf("image/%s", path.Ext(img.LocalPath)))
			ctx.File(img.LocalPath)
			return
		}
		exception.ErrResourceNotFound.ResponseWithError(ctx, "not data")
	}
}
