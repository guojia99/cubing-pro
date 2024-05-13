package posts

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type DeleteTopicReq struct {
	TopicId uint `uri:"topicId"`
}

func DeleteTopic(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req GetTopicReq

		if err := ctx.BindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		if err := svc.DB.Delete(&post.Topic{}, "id = ?", req.TopicId).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
