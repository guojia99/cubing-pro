package posts

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type BanTopicReq struct {
	TopicId uint `uri:"topicId"`
}

func BanTopic(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req GetTopicReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var tp post.Topic
		if err := svc.DB.First(&tp, "id = ?", req.TopicId).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		tp.Ban = true

		if err := svc.DB.Save(&tp); err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, nil)
	}
}
