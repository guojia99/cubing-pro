package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/guojia99/cubing-pro/src/api/app"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/svc"
)

type RegisterReq struct {
	// 用户数据
	LoginID    string `json:"loginID"`    // 登录用的ID
	UserName   string `json:"userName"`   // 用户昵称
	ActualName string `json:"actualName"` // 真实姓名
	EnName     string `json:"enName"`     // 用户英文名
	Password   string `json:"password"`   // 密码（加密后）
	TimeStamp  int64  `json:"timeStamp"`  // 创建时间戳
	Ip         string `json:"ip"`         // ip地址

	// 第三方数据
	QQ    string `json:"QQ"`    // qq号
	Email string `json:"email"` // 邮箱
	WcaID string `json:"WcaID"` // wcaID
	Phone string `json:"phone"` // 手机号

	// 验证码
	VerifyId    string `json:"verifyId"`
	VerifyValue string `json:"verifyValue"`
}

type RegisterResp struct {
	app.GenerallyResp
	Refresh string `json:"refresh"` // 长期刷新秘钥
	Token   string `json:"token"`
	Timeout int64  `json:"timeout"`
}

func Register(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req RegisterReq
		if err := ctx.Bind(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		if ok := verifyCaptcha(req.VerifyId, req.VerifyValue); !ok {
			exception.ErrVerifyCodeField.ResponseWithError(ctx, nil)
			return
		}

		key := utils.GenerateRandomKey(req.TimeStamp)
		password, err := utils.Decrypt(req.Password, key)
		if err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		name := req.UserName
		if req.ActualName != "" {
			name = req.ActualName
		} else if req.EnName != "" {
			name = req.EnName
		}

		var newUser = user.User{
			Name:            req.UserName,
			EnName:          req.EnName,
			LoginID:         req.LoginID,
			ActualName:      req.ActualName,
			Password:        password,
			HistoryPassword: password,
			QQ:              req.QQ,
			Email:           req.Email,
			WcaID:           req.WcaID,
			Phone:           req.Phone,
			Hash:            string(utils.GenerateRandomKey(time.Now().UnixNano())),
			CubeID:          svc.Cov.GetCubeID(name),
			ActivationTime:  time.Now(),
		}

		if err = svc.DB.Create(&newUser).Error; err != nil {
			exception.ErrRegisterField.ResponseWithError(ctx, err)
			return
		}
		ctx.JSON(
			http.StatusOK, RegisterResp{
				Refresh: "",
				Token:   "",
				Timeout: 0,
			},
		)
	}
}
