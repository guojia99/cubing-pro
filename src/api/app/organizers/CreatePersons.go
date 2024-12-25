package organizers

import (
	"fmt"

	email2 "github.com/guojia99/cubing-pro/src/internel/email"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type CreatePersonsReq struct {
	Users []string `json:"Users"`
}

func CreatePersons(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CreatePersonsReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		// 对比
		org := ctx.Value(org_mid.OrgAuthMiddlewareKey).(user.Organizers)
		if len(utils.Has(org.Users(), req.Users)) > 0 {
			exception.ErrResultCreate.ResponseWithError(ctx, "存在已加入的玩家")
			return
		}
		var users []user.User
		svc.DB.First(&users, "cube_id in ?", req.Users)
		if len(users) != len(req.Users) {
			exception.ErrResourceNotFound.ResponseWithError(ctx, "有不存在的用户")
			return
		}

		// 保存
		org.SetUsersCubingID(req.Users)
		if err := svc.DB.Save(&org).Error; err != nil {
			exception.ErrResultCreate.ResponseWithError(ctx, err)
			return
		}

		for _, u := range users {
			_ = email2.SendEmailWithTemp(
				svc.Cfg.GlobalConfig.EmailConfig, "加入主办团队", []string{u.Email}, email2.CodeTemp, email2.CodeTempData{
					Subject:   "加入主办团队",
					UserName:  u.Name,
					Notify:    "加入主办团队",
					NotifyMsg: fmt.Sprintf("你被添加到%s的主办团队成员列表, 你现在可以通过下述链接查看主办团队, 如果这不是你的主办团队, 请及时联系主办 %s 以免造成误解。", org.Name, org.LeaderID),
					NotifyUrl: "", //todo
				},
			)
		}
		exception.ResponseOK(ctx, nil)
	}
}
