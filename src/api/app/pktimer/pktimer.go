package pktimer

import (
	"strings"

	"github.com/gin-gonic/gin"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	pktimerDB "github.com/guojia99/cubing-pro/src/internel/database/model/pktimer"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/robot/types"
)

type PKtimerResponse struct {
	pktimerDB.PkTimerResult

	GroupName string `json:"groupName"`
}

func GetPKtimer(svc *svc.Svc) gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var groups []competition.CompetitionGroup
		svc.DB.Find(&groups)

		var data []pktimerDB.PkTimerResult
		_, _ = app_utils.GenerallyList[pktimerDB.PkTimerResult](
			ctx, svc.DB, data, app_utils.ListSearchParam[pktimerDB.PkTimerResult]{
				Model:   &pktimerDB.PkTimerResult{},
				MaxSize: 20,
				NextFn: func(pk pktimerDB.PkTimerResult) pktimerDB.PkTimerResult {

					pk.PkResults.FirstMessage = types.InMessage{}
					for _, g := range groups {
						if strings.Contains(g.QQGroups, pk.GroupID) || strings.Contains(g.QQGroupUid, pk.GroupID) {
							pk.GroupName = g.Name
							break
						}
					}
					for idx := range pk.PkResults.Players {
						pk.PkResults.Players[idx].QQBot = ""
					}

					return pk
				},
			},
		)
	}
}
