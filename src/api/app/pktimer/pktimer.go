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

// buildGroupMap 构建群组映射表，优化查找性能
func buildGroupMap(groups []competition.CompetitionGroup) map[string]string {
	groupMap := make(map[string]string, len(groups))
	for _, g := range groups {
		// 将 QQGroups 和 QQGroupUid 都映射到群组名称
		if g.QQGroups != "" {
			groupMap[g.QQGroups] = g.Name
		}
		if g.QQGroupUid != "" {
			groupMap[g.QQGroupUid] = g.Name
		}
	}
	return groupMap
}

// findGroupName 查找群组名称，优先使用精确匹配，然后使用包含匹配
func findGroupName(groupID string, groupMap map[string]string, groups []competition.CompetitionGroup) string {
	// 先尝试精确匹配
	if name, ok := groupMap[groupID]; ok {
		return name
	}
	// 回退到包含匹配（兼容旧逻辑）
	for _, g := range groups {
		if strings.Contains(g.QQGroups, groupID) || strings.Contains(g.QQGroupUid, groupID) {
			return g.Name
		}
	}
	return ""
}

func GetPKtimer(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var groups []competition.CompetitionGroup
		if err := svc.DB.Find(&groups).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "查询群组失败"})
			return
		}

		// 构建映射表以提高查找效率
		groupMap := buildGroupMap(groups)

		var data []pktimerDB.PkTimerResult
		_, _ = app_utils.GenerallyList[pktimerDB.PkTimerResult](
			ctx, svc.DB, data, app_utils.ListSearchParam[pktimerDB.PkTimerResult]{
				Model:   &pktimerDB.PkTimerResult{},
				MaxSize: 20,
				NextFn: func(pk pktimerDB.PkTimerResult) pktimerDB.PkTimerResult {
					// 清空敏感信息
					pk.PkResults.FirstMessage = types.InMessage{}
					
					// 查找群组名称
					pk.GroupName = findGroupName(pk.GroupID, groupMap, groups)
					
					// 清空玩家 QQBot 信息
					for idx := range pk.PkResults.Players {
						pk.PkResults.Players[idx].QQBot = ""
					}

					return pk
				},
			},
		)
	}
}
