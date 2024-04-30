package notify

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"

	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type DeleteNotifyReq struct {
	NotifyId uint `uri:"notifyId"`
}

func DeleteNotify(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req DeleteNotifyReq

		if err := ctx.BindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		if err := svc.DB.Delete(&post.Notification{}, "id = ?", req.NotifyId).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, nil)
	}
}
