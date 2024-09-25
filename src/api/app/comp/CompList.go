package comp

import (
	"github.com/gin-gonic/gin"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func List(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var comps []competition.Competition
		app_utils.GenerallyList(
			ctx, svc.DB, comps, app_utils.ListSearchParam{
				Model:   &competition.Competition{},
				MaxSize: 100,
				Query:   "status = ?",
				QueryCons: []interface{}{
					competition.Running,
				},
				CanSearchAndLike: []string{
					"name",
					"id",
					"str_id",
					"country",
					"city",
					"genre",
					"comp_start_time",
					"comp_end_time",
				},
				Select: []string{
					"str_id", "name", "country", "city", "genre",
					"status", "count", "logo",
					"comp_start_time", "comp_end_time",
					"event_min", "series", "wca_url",
				},
			},
		)

	}
}
