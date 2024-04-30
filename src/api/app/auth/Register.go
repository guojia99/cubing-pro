package auth

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/svc"

	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type RegisterReq struct {
	// 用户数据
	LoginID    string `json:"loginID"`    // 登录用的ID
	UserName   string `json:"userName"`   // 用户昵称
	ActualName string `json:"actualName"` // 真实姓名
	EnName     string `json:"enName"`     // 用户英文名
	Password   string `json:"password"`   // 密码（加密后）
	TimeStamp  int64  `json:"timestamp"`  // 创建时间戳

	// 第三方数据
	QQ    string `json:"QQ"`    // qq号
	Email string `json:"email"` // 邮箱
	WcaID string `json:"WcaID"` // wcaID
	Phone string `json:"phone"` // 手机号

	// 验证码
	VerifyId    string `json:"verifyId"`
	VerifyValue string `json:"verifyValue"`
}

func Register(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req RegisterReq
		if err := ctx.Bind(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		//if ok := middleware.Code().VerifyCaptcha(req.VerifyId, req.VerifyValue); !ok {
		//	exception.ErrVerifyCodeField.ResponseWithError(ctx, nil)
		//	return
		//}

		key := utils.GenerateRandomKey(req.TimeStamp)
		fmt.Println(key, req.TimeStamp)
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

		// todo 验证字段， 验证WcaID是否被注册

		if err = svc.DB.Create(&newUser).Error; err != nil {
			exception.ErrRegisterField.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
