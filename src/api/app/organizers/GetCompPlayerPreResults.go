package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func GetCompPlayerPreResult(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		comp := ctx.Value(org_mid.CompMiddlewareKey).(competition.Competition)

		var out []result.PreResults
		app_utils.GenerallyList(
			ctx, svc.DB, out, app_utils.ListSearchParam{
				Model:   &result.PreResults{},
				MaxSize: 100,
				Query:   "comp_id = ?",
				QueryCons: []interface{}{
					comp.ID,
				},
			},
		)

	}
}
