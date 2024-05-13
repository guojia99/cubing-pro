package posts

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type CreateForumReq struct {
	post.Forum
}

func CreateForum(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CreateForumReq
		if err := ctx.ShouldBind(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		if err := svc.DB.Create(&req.Forum).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, nil)
	}
}
