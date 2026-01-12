package notify

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func List(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var find []post.Notification
		app_utils.GenerallyList(
			ctx, svc.DB, find, app_utils.ListSearchParam[post.Notification]{
				Model:   &post.Notification{},
				MaxSize: 20,
				Omit: []string{
					"content",
					"remark",
				},
			},
		)
	}
}
