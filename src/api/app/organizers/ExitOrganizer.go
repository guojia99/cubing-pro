package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func ExitOrganizer(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		org := ctx.Value(org_mid.OrgAuthMiddlewareKey).(user.Organizers)

		userD, err := middleware.GetAuthUser(ctx)
		if err != nil {
			exception.ErrAuthField.ResponseWithError(ctx, err)
			return
		}

		// todo 发邮箱
		if userD.CubeID == org.LeaderID {
			org.Status = user.Disband // 解散但保留所有成员
		} else {
			org.DeleteUserID([]string{userD.CubeID})
		}

		if err = svc.DB.Save(&org).Error; err != nil {
			exception.ErrResultDelete.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
		return
	}
}
