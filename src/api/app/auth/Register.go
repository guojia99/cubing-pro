package auth

import (
	"fmt"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/email"

	"github.com/gin-gonic/gin"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/svc"

	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type PasswordCheckRequest struct {
	Password  string `json:"password"`  // 密码（加密前）
	TimeStamp int64  `json:"timestamp"` // 创建时间戳
}

func PasswordCheck(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req PasswordCheckRequest
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		password, err := utils.EnPwdCode(req.Password, req.TimeStamp)
		if err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, gin.H{"password": password})
	}
}

type RegisterReq struct {
	// 用户数据
	LoginID    string `json:"loginID"`
	UserName   string `json:"userName"`
	ActualName string `json:"actualName"`
	EnName     string `json:"enName"`
	Password   string `json:"password"`
	TimeStamp  int64  `json:"timestamp"`
	QQ         string `json:"QQ"`

	// 验证
	Email     string `json:"email"`
	EmailCode string `json:"emailCode"`

	// 旧数据
	CubeID       string `json:"cubeID"`       // v2 留下的 CubeID
	InitPassword string `json:"initPassword"` // 初始化密码 todo bind
}

func Register(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req RegisterReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		// 验证邮箱验证码数据
		var checker user.CheckCode
		if err := svc.DB.
			Where("`use` = ?", false).
			Where("email = ?", req.Email).
			Where("typ = ?", user.RegisterWithEmail).
			Order("created_at desc").First(&checker).Error; err != nil {
			fmt.Println(err)
			exception.ErrRegisterField.ResponseWithError(ctx, fmt.Errorf("验证码不存在"))
			return
		}

		if time.Since(checker.Timeout) > 0 {
			exception.ErrRegisterField.ResponseWithError(ctx, "邮箱验证码过期")
			return
		}
		if checker.Code != req.EmailCode {
			exception.ErrRegisterField.ResponseWithError(ctx, "邮箱验证码错误")
			return
		}

		// 初始化数据
		name := req.UserName
		if req.ActualName != "" {
			name = req.ActualName
		} else if req.EnName != "" {
			name = req.EnName
		}
		// 验证密码是否有效
		password, err := utils.DePwdCode(req.Password, req.TimeStamp)
		if err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		var newUser user.User
		newUser.SetAuth(user.AuthPlayer)

		if len(req.CubeID) != 0 && len(req.InitPassword) != 0 {
			if err = svc.DB.First(&newUser, "cube_id = ?", req.CubeID).Error; err != nil {
				exception.ErrUserNotFound.ResponseWithError(ctx, err)
				return
			}
			if newUser.InitPassword != req.InitPassword {
				exception.ErrRegisterField.ResponseWithError(ctx, "依据原有用户进行注册初始化密码错误")
				return
			}
			newUser.InitPassword = ""
		}

		if req.QQ != "" {
			newUser.QQ = req.QQ
		}
		if newUser.CubeID == "" {
			newUser.CubeID = svc.Cov.GetCubeID(name)
		}
		newUser.Name = req.UserName
		newUser.EnName = req.EnName
		newUser.ActivationTime = utils.PtrNow()
		newUser.LoginID = req.LoginID
		newUser.ActualName = req.ActualName
		newUser.Password = password
		newUser.HistoryPassword = password
		newUser.Email = req.Email
		newUser.Hash = string(utils.GenerateRandomKey(time.Now().UnixNano()))

		// 创建用户
		newUser.LastUpdateNameTime = utils.PtrNow()
		if err = svc.DB.Save(&newUser).Error; err != nil {
			exception.ErrRegisterField.ResponseWithError(ctx, err)
			return
		}

		svc.DB.Model(&result.Results{}).Where("cube_id = ?", newUser.CubeID).Update("person_name", req.UserName)
		svc.DB.Model(&competition.Registration{}).Where("user_id = ?", newUser.ID).Update("user_name", req.UserName)

		checker.Use = true
		svc.DB.Save(&checker)

		exception.ResponseOK(ctx, nil)
	}
}

type SendRegisterEmailCodeReq struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type SendRegisterEmailCodeResp struct {
	Email        string    `json:"email"`
	Timeout      time.Time `json:"timeout"`
	LastSendTime time.Time `json:"lastSendTime"`
}

func SendRegisterEmailCode(svc *svc.Svc, typ string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req SendRegisterEmailCodeReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var emailUser user.User
		if err := svc.DB.Where("email = ?", req.Email).First(&emailUser).Error; err == nil || emailUser.ID != 0 {
			exception.ErrRegisterField.ResponseWithError(ctx, "该邮箱已被注册，请联系管理员")
			return
		}

		var checker user.CheckCode
		if err := svc.DB.Where("email = ?", req.Email).Where("use = ?", false).Where("typ = ?", typ).Order("created_at desc").First(&checker); err == nil {
			if time.Since(checker.CreatedAt) < time.Minute {
				exception.ResponseOK(
					ctx, SendRegisterEmailCodeResp{
						Email:        req.Email,
						Timeout:      checker.Timeout,
						LastSendTime: checker.CreatedAt.Add(time.Minute),
					},
				)
				return
			}
		}

		checker = user.CheckCode{
			Type:    typ,
			Email:   req.Email,
			Use:     false,
			Code:    utils.RandomString(6),
			Timeout: time.Now().Add(time.Minute * 5),
		}

		subject := "CubingPro"

		switch typ {
		case user.RegisterWithEmail:
			subject += "注册"
		}

		data := email.CodeTempData{
			Subject:        subject,
			UserName:       req.Name,
			Option:         "注册",
			OptionsTimeOut: checker.Timeout.Format(time.DateTime),
			OptionsCode:    checker.Code,
		}

		if err := email.SendEmailWithTemp(svc.Cfg.GlobalConfig.EmailConfig, subject, []string{req.Email}, email.CodeTemp, data); err != nil {
			exception.ErrRegisterField.ResponseWithError(ctx, err)
			return
		}
		_ = svc.DB.Save(&checker)
		exception.ResponseOK(
			ctx, SendRegisterEmailCodeResp{
				Email:        req.Email,
				Timeout:      checker.Timeout,
				LastSendTime: time.Now().Add(time.Minute),
			},
		)
	}
}
