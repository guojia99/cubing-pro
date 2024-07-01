package posts

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type CreateTopicReq struct {
	ForumID    uint   `json:"ForumID"`                          // 板块ID
	Title      string `json:"Title" binding:"required,max=100"` // 标题
	Short      string `json:"Short" binding:"max=300"`          // 简短说明
	Content    string `json:"Content"`                          // md
	Tags       string `json:"Tags"`                             // tags
	Type       string `json:"Type"`                             // 类型
	TopImage   string `json:"TopImage"`                         // 头图
	IsOriginal bool   `json:"IsOriginal"`                       // 是否原创
	Original   string `json:"Original"`                         // 原创
	KeyWords   string `json:"KeyWords"`                         // 关键词
}

func CreateTopic(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}

		var req CreateTopicReq
		if err = app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		topic := post.Topic{
			Fid:            req.ForumID,
			CreateBy:       user.Name,
			CreateByUserID: user.ID,
			CreateIp:       ctx.ClientIP(),
			UpdateIp:       ctx.ClientIP(),
			Status:         post.TopicStatusUnpublished,
			Title:          req.Title,
			Short:          req.Short,
			Content:        req.Content,
			Tags:           req.Tags,
			Type:           req.Type,
			TopImage:       req.TopImage,
			IsOriginal:     req.IsOriginal,
			Original:       req.Original,
			KeyWords:       req.KeyWords,
			Ban:            false,
		}

		if err = svc.DB.Create(&topic).Error; err != nil {
			exception.ErrResultCreate.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(
			ctx, gin.H{
				"id": topic.ID,
			},
		)
	}
}

func ReleaseTopic(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}

		var req GetTopicReq
		if err = app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var topic post.Topic
		if err = svc.DB.First(&topic, "id = ?", req.TopicId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		if !(topic.Status == post.TopicStatusUnpublished || topic.Status == post.TopicStatusReviewField) {
			exception.ErrResultUpdate.ResponseWithError(ctx, "状态需为未发布")
			return
		}

		topic.Status = post.TopicStatusPendingReview
		topic.UpdateIp = ctx.ClientIP()
		svc.DB.Save(&topic)
		exception.ResponseOK(ctx, nil)
	}
}
