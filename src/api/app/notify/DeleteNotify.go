package notify

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"

	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type DeleteNotifyReq struct {
	NotifyId uint `uri:"notifyId"`
}

func DeleteNotify(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req DeleteNotifyReq

		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		if err := svc.DB.Delete(&post.Notification{}, "id = ?", req.NotifyId).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, nil)
	}
}
