package sports

import (
	"github.com/gin-gonic/gin"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/sports"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func ListSportResults(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var results []sports.SportResult

		app_utils.GenerallyList(
			ctx, svc.DB, results, app_utils.ListSearchParam[sports.SportResult]{
				Model:   &sports.SportResult{},
				MaxSize: 100,
				CanSearchAndLike: []string{
					"event_id", "user_id", "date",
				},
			},
		)
	}
}
