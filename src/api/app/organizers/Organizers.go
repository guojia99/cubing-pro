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

type MeOrganizersReq struct {
	Status user2.OrganizersStatus `query:"Status"`
}

func MeOrganizers(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req MeOrganizersReq
		if err := ctx.ShouldBindQuery(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			exception.ErrAuthField.ResponseWithError(ctx, err)
			return
		}

		var list []user2.Organizers
		utils.GenerallyList(
			ctx, svc.DB, list, utils.ListSearchParam{
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
