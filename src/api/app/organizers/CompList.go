package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func OrgCompList(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		org := ctx.Value(OrgAuthMiddlewareKey).(user.Organizers)
		var list []competition.Competition
		utils.GenerallyList(
			ctx, svc.DB, list, utils.ListSearchParam{
				Model:   &competition.Competition{},
				MaxSize: 100,
				Query:   "orgId = ?",
				QueryCons: []interface{}{
					org.ID,
				},
				Select: []string{
					"str_id", "name", "city",
					"genre", "count", "free_p",
					"comp_start_time", "comp_end_time",
					"reg_start_time", "reg_end_time",
					"reg_cancel_dl_time", "reg_restart_time",
					"orgId", "wca_url", "min_count", "count",
					"status", "reject_msg",
				},
				HasDeleted: true,
			},
		)
	}
}

//E B Z
//EK	R:[R D' R',U]	KE	R:[U,R D' R']
//EG	R:[R D R',U]	GE	R:[U,R D R']
//EW	R:[R D2 R',U]	WE	R:[U,R D2 R']
//BK	R:[R D' R',U2]	KB	R:[U2,R D' R']
//BG	R:[R D R',U2]	GB	R:[U2,R D R']
//BW	R:[R D2 R',U2]	WB	R:[U2,R D2 R']
//ZK	R:[R D' R',U']	KZ	R:[U',R D' R']
//ZW	R:[R D2 R',U']	WZ	R:[U',R D2 R']
//ZG	R:[R D R',U']	GZ	R:[U',R D R']
