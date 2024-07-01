package users

import (
	"github.com/gin-gonic/gin"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

// todo 热门查询
func Users(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var out []user.User
		app_utils.GenerallyList(
			ctx, svc.DB, out, app_utils.ListSearchParam{
				Model:   &user.User{},
				MaxSize: 100,
				CanSearchAndLike: []string{
					"cube_id", "en_name", "name",
				},
				Query:     "ban = ?",
				QueryCons: []interface{}{false},
				Select: []string{
					"id", "name", "en_name", "cube_id", "represent_name",
				},
			},
		)
	}
}
