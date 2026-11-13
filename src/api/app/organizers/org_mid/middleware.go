package org_mid

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	user2 "github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"gorm.io/gorm"
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

// Admin routes already enforce JWT and admin roles; resolve the actual owner here.
func AdminCompMiddleware(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CheckCompMiddlewareReq
		if err := ctx.ShouldBindUri(&req); err != nil || req.CompId == 0 {
			exception.ErrRequestBinding.ResponseWithError(ctx, "无效比赛ID")
			ctx.Abort()
			return
		}
		var c competition.Competition
		if err := svc.DB.First(&c, "id = ?", req.CompId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			ctx.Abort()
			return
		}
		// DELETE existed for all competition types in HEAD; new management APIs are group-only.
		if !c.IsGroupCompetition() && ctx.Request.Method != "DELETE" {
			exception.ErrResourceNotFound.ResponseWithError(ctx, "该管理接口仅支持群赛")
			return
		}
		var org user2.Organizers
		if err := svc.DB.First(&org, "id = ?", c.OrganizersID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) && (ctx.Request.Method == "GET" || ctx.Request.Method == "DELETE") {
				org.ID = c.OrganizersID
			} else {
				exception.ErrResourceNotFound.ResponseWithError(ctx, err)
				ctx.Abort()
				return
			}
		}
		ctx.Set(CompMiddlewareKey, c)
		ctx.Set(OrgAuthMiddlewareKey, org)
		ctx.Next()
	}
}
