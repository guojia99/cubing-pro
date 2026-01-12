package posts

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func GetAllTopics(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var posts []post.Topic
		app_utils.GenerallyList(
			ctx, svc.DB, posts, app_utils.ListSearchParam[post.Topic]{
				Model:      &post.Topic{},
				MaxSize:    100,
				HasDeleted: true,
				Omit: []string{
					"content",
				},
			},
		)
	}
}
