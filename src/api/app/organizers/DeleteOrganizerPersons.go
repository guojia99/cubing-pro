package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type DeletePersonsReq struct {
	PersonId string `uri:"personId"`
}

func DeletePersons(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req DeletePersonsReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		org := ctx.Value(org_mid.OrgAuthMiddlewareKey).(user.Organizers)
		org.DeleteUserID([]string{req.PersonId})

		// todo 发邮件
		if err := svc.DB.Save(&org).Error; err != nil {
			exception.ErrResultDelete.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, nil)
	}
}
