package posts

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	"github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type GetPostsReq struct {
	TopicId uint `uri:"topicId"`
}

func GetPosts(svc *svc.Svc, delete bool, user bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var req GetPostsReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var query []string
		var queryCons []interface{}

		if req.TopicId != 0 {
			query = append(query, "tid = ?")
			queryCons = append(queryCons, req.TopicId)
		}
		if user {
			usr, err := middleware.GetAuthUser(ctx)
			if err != nil {
				return
			}
			query = append(query, "uid = ?")
			queryCons = append(queryCons, usr.ID)
		}

		var posts []post.Posts
		app_utils.GenerallyList(
			ctx, svc.DB, posts, app_utils.ListSearchParam[post.Posts]{
				Model:      &post.Posts{},
				MaxSize:    100,
				Query:      strings.Join(query, "AND"),
				QueryCons:  queryCons,
				HasDeleted: delete,
			},
		)
	}
}
