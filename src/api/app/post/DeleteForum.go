package posts

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type DeleteForumReq struct {
	ForumId string `uri:"forumId"`
}

func DeleteForum(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req DeleteForumReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		if err := svc.DB.Delete(&post.Forum{}, "id = ?", req.ForumId).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
