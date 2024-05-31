package posts

import (
	"github.com/gin-gonic/gin"
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
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var posts []post.Posts
		app_utils.GenerallyList(
			ctx, svc.DB, posts, app_utils.ListSearchParam{
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

//2. U B D2 B' U R2 B U2 R2 B U2 B' L2 F R2 L2 D2
//3. F2 B' U2 L' U2 L B U2 D2 L2 U2 B2 L B2 L D2 F2 R2
//4. U2 L D2 F2 U2 B2 R' U B2 U L2 F2 D2 L2 U F2 U D2
//5. R2 D2 B2 U' F2 R2 U' L2 B2 D2 U' L2 B' R2 D' B' R2 U' B'
//6. B D2 L2 U2 B L2 B D2 B D2 B2 R2 D B R2 B' D F R2 D
//7. B2 D F2 U' L2 U2 B2 R2 F2 U B2 R2 F' L2 D L2 D' F' L2 U'
//8. U2 B L2 U2 L2 U2 B U' F2 D' B2 U L2 D' B2 L2 U2 F2 R2
//9. F' L2 B R2 U2 F' R2 U2 F U2 L2 F2 D' B' R2 B R2 D' F2
//10. L2 U R2 U' L2 B2 U B2 U R2 B2 R U R' U B2 R U2 R'
