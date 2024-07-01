package result

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func PreResults(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}

		var out []result.PreResults
		app_utils.GenerallyList(
			ctx, svc.DB, out, app_utils.ListSearchParam{
				Model:     &result.PreResults{},
				MaxSize:   100,
				Query:     "user_id = ?",
				QueryCons: []interface{}{user.ID},
			},
		)
	}
}
