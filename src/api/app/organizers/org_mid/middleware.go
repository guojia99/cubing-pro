package org_mid

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	user2 "github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

const OrgAuthMiddlewareKey = "organizerID"
const CompMiddlewareKey = "compID"

type OrgAuthMiddlewareReq struct {
	OrgId uint `uri:"orgId"`
}

func OrgAuthMiddleware(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req OrgAuthMiddlewareReq

		if err := ctx.BindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}

		var org user2.Organizers
		err = svc.DB.
			Where("leaderId = ? OR ass_org_users like ?", user.CubeID, fmt.Sprintf("%%%s%%", user.CubeID)).
			Where("id = ?", req.OrgId).
			First(&org).Error

		if err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		ctx.Set(OrgAuthMiddlewareKey, org)
		ctx.Next()
	}
}

func CheckOrgCanUse() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		org := ctx.Value(OrgAuthMiddlewareKey).(user2.Organizers)

		if org.CanUse() {
			ctx.Next()
			return
		}

		exception.ErrResultCanNotUse.ResponseWithError(ctx, "不可操作状态")
	}
}

type CheckCompMiddlewareReq struct {
	CompId uint `uri:"compId" json:"compId"`
}

func CheckCompMiddleware(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CheckCompMiddlewareReq
		if err := ctx.BindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		var comp competition.Competition
		if err := svc.DB.First(&comp, "id = ?", req.CompId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		org := ctx.Value(OrgAuthMiddlewareKey).(user2.Organizers)
		if comp.OrganizersID != org.ID {
			exception.ErrAuthField.ResponseWithError(ctx, "无权限")
			return
		}
		ctx.Set(CompMiddlewareKey, comp)
		ctx.Next()
	}
}
