package organizers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	"github.com/guojia99/cubing-pro/src/api/utils"
	user2 "github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type OrganizersReq struct {
	Status user2.OrganizersStatus `query:"Status"`
}

func MeOrganizers(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req OrganizersReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			exception.ErrAuthField.ResponseWithError(ctx, err)
			return
		}

		var list []user2.Organizers
		app_utils.GenerallyList(
			ctx, svc.DB, list, app_utils.ListSearchParam{
				Model:   &user2.Organizers{},
				MaxSize: 100,
				Query:   "leaderId = ? OR ass_org_users like ?",
				QueryCons: []interface{}{
					user.CubeID,
					fmt.Sprintf("%%%s%%", user.CubeID),
				},

				HasDeleted: false,
			},
		)
	}
}

func AllOrganizers(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req OrganizersReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var list []user2.Organizers
		app_utils.GenerallyList(
			ctx, svc.DB, list, app_utils.ListSearchParam{
				Model:      &user2.Organizers{},
				MaxSize:    100,
				HasDeleted: false,
			},
		)
	}
}
