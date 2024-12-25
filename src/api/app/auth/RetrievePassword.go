package auth

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/email"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	user2 "github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type RetrievePasswordSendCodeReq struct {
	LoginID string `json:"loginID"`
	Type    string `json:"type"` // email | phone(短信验证)
}

type RetrievePasswordResp struct {
	Key      string    `json:"key"`      // 授权key
	Type     string    `json:"type"`     // 处理类型
	Email    string    `json:"email"`    // 掩码处理
	Phone    string    `json:"phone"`    // 掩码处理
	Timeout  time.Time `json:"timeout"`  // 超时时间
	LastSend time.Time `json:"lastSend"` // 下次发送时间
	Msg      string    `json:"msg"`
}

func RetrievePasswordSendCode(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req RetrievePasswordSendCodeReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		// 查询用户
		var user user2.User
		sql := svc.DB.Where("login_id = ?", req.LoginID)
		if utils.IsEmailValid(req.LoginID) {
			sql = svc.DB.Where("email = ?", req.LoginID)
		}
		if utils.IsPhoneNumberValid(req.LoginID) {
			sql = svc.DB.Where("phone = ?", req.LoginID)
		}
		if err := sql.First(&user).Error; err != nil {
			exception.ErrUserNotFound.ResponseWithError(ctx, err)
			return
		}

		// 防止反复执行
		var checker user2.CheckCode
		if err := svc.DB.Where("use = ?", false).Where("uid = ?", user.ID).Where("typ = ?", req.Type).Order("created_at desc").First(&checker); err == nil {
			if time.Since(checker.CreatedAt) < time.Minute {
				exception.ResponseOK(
					ctx, RetrievePasswordResp{
						Key:      checker.Key,
						Type:     req.Type,
						Email:    utils.HideEmail(user.Email),
						Timeout:  checker.Timeout,
						LastSend: checker.CreatedAt.Add(time.Minute),
						Msg:      "请勿反复请求",
					},
				)
				return
			}
		}

		// 新建变量
		checker = user2.CheckCode{
			Type:    req.Type,
			UserID:  user.ID,
			Use:     false,
			Key:     utils.RandomString(32),
			Code:    strings.ToUpper(utils.RandomString(32)),
			Timeout: time.Now().Add(time.Minute * 5),
		}
		var resp = RetrievePasswordResp{
			Key:      checker.Key,
			Type:     req.Type,
			Email:    "",
			Phone:    "",
			Timeout:  checker.Timeout,
			LastSend: checker.CreatedAt.Add(time.Minute),
			Msg:      "创建成功",
		}

		// 分发验证码
		switch req.Type {
		case user2.RetrievePasswordWithEmail:
			if user.Email == "" {
				exception.ErrResourceNotFound.ResponseWithError(ctx, "该用户无email, 请使用其他方法进行找回密码")
				return
			}
			resp.Email = utils.HideEmail(user.Email)

			urlP, err := url.JoinPath(svc.Cfg.GlobalConfig.BaseHost, "/retrieve")
			if err != nil {
				exception.ErrInternalServer.ResponseWithError(ctx, err)
				return
			}
			urlP += fmt.Sprintf(
				"?key=%s&code=%s&email=%s&ts=%d&timeout=%d&uid=%d&typ=%s",
				checker.Key, checker.Code, user.Email, time.Now().Unix(), checker.Timeout.Unix(), user.ID, checker.Type,
			)

			data := email.CodeTempData{
				Subject:        "CubePro找回密码",
				UserName:       user.Name,
				Option:         "找回密码",
				OptionsTimeOut: resp.Timeout.Format(time.DateTime),
				OptionsUrl:     urlP, // todo
			}
			if err = email.SendEmailWithTemp(svc.Cfg.GlobalConfig.EmailConfig, data.Subject, []string{user.Email}, email.CodeTemp, data); err != nil {
				exception.ErrAuthField.ResponseWithError(ctx, err)
				return
			}
		default:
			exception.ErrAuthField.ResponseWithError(ctx, fmt.Sprintf("无法使用%s方式进行找回密码", req.Type))
			return
		}

		// 写入数据库
		_ = svc.DB.Create(&checker)
		exception.ResponseOK(ctx, resp)
	}
}

type RetrievePasswordReq struct {
	CheckCodeReq

	Password  string `json:"password"`  // 密码（加密前）
	TimeStamp int64  `json:"timestamp"` // 创建时间戳
}

func RetrievePassword(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req RetrievePasswordReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		// 验证密码是否有效
		password, err := utils.DePwdCode(req.Password, req.TimeStamp)
		if err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		// 验证验证码
		var checker user2.CheckCode
		err = svc.DB.Where("use = ?", false).
			Where("key = ?", req.Key).
			Where("uid = ?", req.UserID).
			Where("email = ?", req.Email).
			Where("code = ?", req.EmailCode).
			Where("typ = ?", req.Type).First(&checker).Error

		if err != nil {
			exception.ErrAuthField.ResponseWithError(ctx, err)
			return
		}

		if time.Since(checker.Timeout) > 0 {
			exception.ErrVerifyCodeField.ResponseWithError(ctx, "验证码过期")
			return
		}

		var user user2.User
		if err = svc.DB.First(&user, "id = ?", req.UserID).Error; err != nil {
			exception.ErrUserNotFound.ResponseWithError(ctx, err)
			return
		}

		user.Password = password
		user.PassWordLockTime = nil
		user.InitPassword = ""
		user.SumPasswordWrong = 0

		if err = svc.DB.Save(&user).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}

		checker.Use = true
		svc.DB.Save(&checker)
		exception.ResponseOK(ctx, "ok")
	}
}
