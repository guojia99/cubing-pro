package posts

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type DeletePostReq struct {
	TopicId uint `uri:"topicId"`
	PostId  uint `uri:"postId"`
}

func DeletePost(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var req DeletePostReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		var p post.Posts
		if err := svc.DB.Where("tid = ?", req.TopicId).Where("id = ?", req.PostId).First(&p).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
