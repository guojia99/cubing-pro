package organizers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type CreateCompReq struct {
	competition.Competition

	Apply bool `json:"Apply"`
}

func CreateComp(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		org := ctx.Value(OrgAuthMiddlewareKey).(user.Organizers)

		var req CreateCompReq
		if err := ctx.ShouldBind(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		fmt.Println(req.Apply)
		// 处理状态
		req.Competition.OrganizersID = org.ID
		req.Competition.Status = competition.Temporary
		if req.Apply {
			req.Competition.Status = competition.Reviewing
			// todo 发邮件
		}

		// 其他
		if req.Competition.StrId == "" {
			req.Competition.StrId = utils.RandomString(32)
		}

		if err := svc.DB.Create(&req.Competition).Error; err != nil {
			exception.ErrResultCreate.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}

type ApplyCompReq struct {
	CompId uint `uri:"CompId"`
}

func ApplyComp(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req ApplyCompReq

		if err := ctx.BindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		org := ctx.Value(OrgAuthMiddlewareKey).(user.Organizers)
		var comp competition.Competition
		if err := svc.DB.First(&comp, "id = ? and orgId = ?", req.CompId, org.ID).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		switch comp.Status {
		case competition.Reject, competition.Temporary:
		default:
			exception.ErrResultUpdate.ResponseWithError(ctx, "该状态无法审批")
			return
		}

		comp.Status = competition.Reviewing
		if err := svc.DB.Save(&comp); err != nil {
			exception.ErrResultUpdate.ResponseWithError(ctx, err)
			return
		}
		// todo 发邮件
		exception.ResponseOK(ctx, nil)
	}
}
