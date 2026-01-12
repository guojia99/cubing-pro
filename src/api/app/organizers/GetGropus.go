package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func GetGroups(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		org := ctx.Value(org_mid.OrgAuthMiddlewareKey).(user.Organizers)
		var groups []competition.CompetitionGroup
		_, _ = app_utils.GenerallyList(
			ctx, svc.DB, groups, app_utils.ListSearchParam[competition.CompetitionGroup]{
				Model:     &competition.CompetitionGroup{},
				MaxSize:   0,
				Query:     "orgId = ?",
				QueryCons: []interface{}{org.ID},
			})
	}
}
