package organizers

import (
	"fmt"

	"github.com/guojia99/cubing-pro/src/api/public"

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

type MeOrganizersData struct {
	user2.Organizers
	Users []public.User
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
		dest, err := app_utils.GenerallyList(
			ctx, svc.DB, list, app_utils.ListSearchParam{
				Model:   &user2.Organizers{},
				MaxSize: 0,
				Query:   "leaderId = ? OR ass_org_users like ?",
				QueryCons: []interface{}{
					user.CubeID,
					fmt.Sprintf("%%%s%%", user.CubeID),
				},
				HasDeleted:  false,
				NotAutoResp: true,
			},
		)

		// 获取所有用户
		var usersKey []string
		for _, o := range dest {
			usersKey = append(usersKey, o.Users()...)
		}
		var users []user2.User
		svc.DB.Find(&users, "cube_id in ?", usersKey)
		var usersMap = make(map[string]user2.User)
		for _, u := range users {
			usersMap[u.CubeID] = u
		}

		var out []MeOrganizersData
		for _, o := range dest {
			m := MeOrganizersData{
				Organizers: o,
				Users:      []public.User{},
			}
			for _, u := range o.Users() {
				usr, ok := usersMap[u]
				if ok {
					m.Users = append(m.Users, public.UserToUser(usr))
				}
			}
			out = append(out, m)
		}

		exception.ResponseOK(
			ctx, app_utils.GenerallyListResp{
				Items: out,
				Total: int64(len(out)),
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
