package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func OrgCompList(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		org := ctx.Value(org_mid.OrgAuthMiddlewareKey).(user.Organizers)
		var list []competition.Competition
		_, _ = app_utils.GenerallyList(
			ctx, svc.DB, list, app_utils.ListSearchParam{
				Model:   &competition.Competition{},
				MaxSize: 0,
				Query:   "orgId = ?",
				QueryCons: []interface{}{
					org.ID,
				},
				Select: []string{
					"str_id", "name", "city", "event_min",
					"genre", "count", "free_p",
					"comp_start_time", "comp_end_time",
					"reg_start_time", "reg_end_time",
					"reg_cancel_dl_time", "reg_restart_time",
					"orgId", "wca_url", "min_count", "count",
					"status", "reject_msg", "is_done",
				},
				HasDeleted: true,
			},
		)
	}
}

type CompsReq struct {
	Status competition.CompetitionStatus `json:"Status"`
}

func Comps(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CompsReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		if req.Status == "" {
			req.Status = competition.Running
		}

		var list []competition.Competition
		app_utils.GenerallyList(
			ctx, svc.DB, list, app_utils.ListSearchParam{
				Model:   &competition.Competition{},
				MaxSize: 100,
				Query:   "status = ?",
				QueryCons: []interface{}{
					req.Status,
				},
			},
		)
	}
}
