package notify

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type GetNotifyListResp struct {
	Tops    []post.Notification `json:"Tops"`
	Results []post.Notification `json:"Results"`
}

func GetNotifyList(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var tops []post.Notification
		if err := svc.DB.Limit(3).Find(&tops, "top = ?", true).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		var results []post.Notification
		app_utils.GenerallyList(
			ctx, svc.DB, results, app_utils.ListSearchParam{
				Model:       &post.Notification{},
				MaxSize:     10,
				NotAutoResp: true,
			},
		)

		exception.ResponseOK(
			ctx, GetNotifyListResp{
				Tops:    tops,
				Results: results,
			},
		)

	}
}
