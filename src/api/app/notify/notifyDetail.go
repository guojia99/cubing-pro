package notify

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type NotifyDetailReq struct {
	NotifyId uint `uri:"notifyId"`
}

func NotifyDetail(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req NotifyDetailReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var notify post.Notification
		if err := svc.DB.First(&notify, "id = ?", req.NotifyId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, notify)
	}
}
