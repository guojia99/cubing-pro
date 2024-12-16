package comp

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type CompReq struct {
	CompId uint `uri:"compId"`
}

type CompResp struct {
	competition.Competition
	Org user.Organizers `json:"Org"`

	Group competition.CompetitionGroup `json:"Group"`

	RegisterNum uint `json:"RegisterNum,omitempty"` // 已注册人数
	CompedNum   uint `json:"CompedNum,omitempty"`   // 已参赛人数

	EarliestID   uint   `json:"EarliestID"`
	EarliestName string `json:"EarliestName"`

	LatestID   uint   `json:"LatestID"`
	LatestName string `json:"LatestName"`
}

func Comp(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CompReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var comp competition.Competition
		if err := svc.DB.First(&comp, "id = ?", req.CompId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		var org user.Organizers
		svc.DB.Omit("leaderId", "ass_org_users", "status", "leader_remark", "admin_msg").
			First(&org, "id = ?", comp.OrganizersID)

		var group competition.CompetitionGroup
		svc.DB.First(&group, "id = ?", comp.GroupID)

		var RegisterNum int64
		svc.DB.Model(&competition.Registration{}).Where("comp_id = ?", comp.ID).Count(&RegisterNum)
		var CompedNum int64
		svc.DB.Model(&result.Results{}).Distinct("user_id").Where("comp_id = ?", comp.ID).Count(&CompedNum)

		var earliestID, latestID uint
		var earliestName, latestName string
		svc.DB.Model(&competition.Competition{}).Where("created_at < ? and status = ?", comp.CreatedAt, competition.Running).Select("id").
			Order("created_at DESC").Limit(1).Pluck("id", &earliestID)
		svc.DB.Model(&competition.Competition{}).Where("created_at < ?  and status = ?", comp.CreatedAt, competition.Running).Select("name").
			Order("created_at DESC").Limit(1).Pluck("name", &earliestName)

		svc.DB.Model(&competition.Competition{}).Where("created_at > ?  and status = ?", comp.CreatedAt, competition.Running).Select("id").
			Order("created_at ASC").Limit(1).Pluck("id", &latestID)
		svc.DB.Model(&competition.Competition{}).Where("created_at > ?  and status = ?", comp.CreatedAt, competition.Running).Select("name").
			Order("created_at ASC").Limit(1).Pluck("name", &latestName)

		// 隐私项目
		comp.RejectMsg = ""
		exception.ResponseOK(
			ctx, CompResp{
				Competition:  comp,
				Org:          org,
				Group:        group,
				RegisterNum:  uint(RegisterNum),
				CompedNum:    uint(CompedNum),
				LatestID:     latestID,
				LatestName:   latestName,
				EarliestID:   earliestID,
				EarliestName: earliestName,
			},
		)
	}
}
