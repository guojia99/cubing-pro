package posts

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type CreatePostReq struct {
	GetTopicReq

	Content  string `json:"Content" binding:"max=400"`
	ReplyPid uint   `uri:"post_id"` // 回复的帖子
}

func CreatePost(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}

		var req CreatePostReq
		if err = app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var replyPost post.Posts
		if req.ReplyPid != 0 {
			if err = svc.DB.First(&replyPost, "id = ?", req.ReplyPid).Error; err != nil {
				exception.ErrResourceNotFound.ResponseWithError(ctx, err)
				return
			}
		}

		var topic post.Topic
		if err = svc.DB.First(&topic, "id = ?", req.TopicId).Error; err != nil {
			exception.ErrUserNotFound.ResponseWithError(ctx, err)
			return
		}

		var p = post.Posts{
			Tid:      req.TopicId,
			Uid:      user.ID,
			UserName: user.Name,
			ReplyPid: replyPost.ID,
			ToName:   replyPost.UserName,
			ToId:     replyPost.Uid,
			Content:  req.Content,
			IP:       ctx.ClientIP(),
		}

		if err = svc.DB.Create(&p).Error; err != nil {
			exception.ErrResultCreate.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)

	}
}
