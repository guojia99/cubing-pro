package comp

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
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

	RegisterNum uint `json:"RegisterNum,omitempty"` // 已注册人数
	CompedNum   uint `json:"CompedNum,omitempty"`   // 已参赛人数
}

func Comp(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CompReq
		if err := ctx.BindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
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

		var RegisterNum int64
		svc.DB.Model(&competition.Registration{}).Where("comp_id = ?", comp.ID).Count(&RegisterNum)
		var CompedNum int64
		svc.DB.Model(&result.Results{}).Distinct("user_id").Where("comp_id = ?", comp.ID).Count(&CompedNum)

		// 隐私项目
		comp.RejectMsg = ""
		exception.ResponseOK(
			ctx, CompResp{
				Competition: comp,
				Org:         org,
				RegisterNum: uint(RegisterNum),
			},
		)
	}
}
