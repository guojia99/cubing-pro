package organizers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/email"
	user2 "github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type RegisterOrganizersReq struct {
	Name         string   `json:"name"`
	Introduction string   `json:"introduction"` // md
	Email        string   `json:"email"`
	QQGroup      []string `gorm:"column:qq_group"`
	QQGroupUid   []string `gorm:"column:qq_group_uid"`
	LeaderRemark string   `gorm:"column:leader_remark"`
}

const RegisterOrganizersEmailMsgT = `
已收到你创建 <p style="display: inline-block;color:red;"> %s </p> 团队的申请，请耐心等待管理员审核。
我们将在1-3个工作日内完成对主办团队的申请审核。你可以在主办界面查看你的团队申请进度，在申请通过后，我们会及时通过邮件给予你通知，请耐心关注邮箱信息。
`

func RegisterOrganizers(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req RegisterOrganizersReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		// 确认用户信息合法性
		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			exception.ErrAuthField.ResponseWithError(ctx, err)
			return
		}
		if user.Email == "" {
			exception.ErrUserResultField.ResponseWithError(ctx, "用户未注册邮箱信息, 请补充邮箱后再申请")
			return
		}

		// todo 有重复的申请？ 一个人最多几个主办团队？
		var org user2.Organizers
		if err = svc.DB.First(&org, "name = ?", req.Name).Error; err == nil {
			exception.ErrResultBeUse.ResponseWithError(ctx, "该团队名字已存在")
			return
		}

		// 创建
		org = user2.Organizers{
			Name:         req.Name,
			Introduction: req.Introduction,
			Email:        req.Email,
			QQGroup:      utils.ToJSON(req.QQGroup),
			QQGroupUid:   utils.ToJSON(req.QQGroupUid),
			LeaderID:     user.CubeID,
			Status:       user2.Applying,
			LeaderRemark: req.LeaderRemark,
		}
		if err = svc.DB.Create(&org).Error; err != nil {
			exception.ErrResultCreate.ResponseWithError(ctx, err)
			return
		}

		// 发送通知邮箱
		data := email.CodeTempData{
			Subject:   "创建主办团队",
			UserName:  user.Name,
			BaseUrl:   svc.Cfg.GlobalConfig.BaseHost,
			Notify:    "创建主办团队申请",
			NotifyMsg: fmt.Sprintf(RegisterOrganizersEmailMsgT, req.Name),
			NotifyUrl: "",
		}
		_ = email.SendEmailWithTemp(svc.Cfg.GlobalConfig.EmailConfig, "创建主办团队", []string{user.Email}, email.CodeTemp, data)

		exception.ResponseOK(ctx, nil)

		// todo 发送邮箱给管理员审批, 审批结束添加主办身份
	}
}
