package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/public"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func Persons(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		org := ctx.Value(org_mid.OrgAuthMiddlewareKey).(user.Organizers)

		var users []user.User
		var out []public.User

		if err := svc.DB.Where("cube_id IN ?", org.Users()).Find(&users).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		for _, val := range users {
			out = append(out, public.UserToUser(val))
		}
		exception.ResponseOK(ctx, out)
	}
}
