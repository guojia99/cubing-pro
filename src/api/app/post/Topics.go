package posts

import (
	"github.com/gin-gonic/gin"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type GetTopicsReq struct {
}

func GetTopics(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var out []post.Topic

		app_utils.GenerallyList(
			ctx, svc.DB, &out, app_utils.ListSearchParam{
				Model:   &post.Topic{},
				MaxSize: 20,
				Query:   "ban = ?",
				QueryCons: []interface{}{
					false,
				},
				CanSearchAndLike: []string{
					"fid", "title", "short", "create_by",
				},
				Omit: []string{
					"content",
				},
			},
		)
	}
}
