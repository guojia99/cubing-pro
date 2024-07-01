package posts

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type DeleteTopicReq struct {
	TopicId uint `uri:"topicId"`
}

func DeleteTopic(svc *svc.Svc, checkUser bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req GetTopicReq

		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var topic post.Topic
		if err := svc.DB.Where("id = ?", req.TopicId).First(&topic).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		if checkUser {
			user, err := middleware.GetAuthUser(ctx)
			if err != nil {
				return
			}
			if topic.CreateByUserID != user.ID {
				exception.ErrAuthField.ResponseWithError(ctx, "无法删除非自己的帖子")
				return
			}
		}

		if err := svc.DB.Delete(&post.Topic{}, "id = ?", req.TopicId).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
