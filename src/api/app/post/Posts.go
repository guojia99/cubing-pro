package posts

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type GetPostsReq struct {
	TopicId uint `uri:"topicId"`
}

func GetPosts(svc *svc.Svc, delete bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var req GetPostsReq
		if err := ctx.BindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		var posts []post.Posts
		utils.GenerallyList(
			ctx, svc.DB, posts, utils.ListSearchParam{
				Model:   &post.Posts{},
				MaxSize: 100,
				Query:   "tid = ?",
				QueryCons: []interface{}{
					req.TopicId,
				},
				HasDeleted: delete,
			},
		)
	}
}
