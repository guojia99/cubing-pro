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
		utils.GenerallyList(ctx, svc.DB, &post.Notification{}, find, 20)
	}
}