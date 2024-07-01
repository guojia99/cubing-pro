package posts

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type GetTopicReq struct {
	TopicId uint `uri:"topicId"`
}

func GetTopic(svc *svc.Svc, ignoreBan bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req GetTopicReq

		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var out post.Topic
		if err := svc.DB.First(&out, "id = ?", req.TopicId).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}

		if ignoreBan && out.Ban {
			exception.ErrDatabase.ResponseWithError(ctx, "帖子已被封禁，无权查看")
			return
		}

		exception.ResponseOK(ctx, out)
	}
}
