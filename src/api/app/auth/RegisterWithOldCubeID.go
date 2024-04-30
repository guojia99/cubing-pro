package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/svc"

	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type RegisterWithOldCubeIDReq struct {
	RegisterReq

	CubeID       string `json:"cubeID"`       // 登录ID
	InitPassword string `json:"initPassword"` // 初始化密码
}

// RegisterWithOldCubeID 依据旧的CubeID进行注册
func RegisterWithOldCubeID(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req RegisterWithOldCubeIDReq
		if err := ctx.Bind(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		if ok := middleware.Code().VerifyCaptcha(req.VerifyId, req.VerifyValue); !ok {
			exception.ErrVerifyCodeField.ResponseWithError(ctx, nil)
			return
		}

		if req.CubeID == "" {
			exception.ErrRequestBinding.ResponseWithError(ctx, fmt.Errorf("cubeId无效"))
			return
		}

		key := utils.GenerateRandomKey(req.TimeStamp)
		password, err := utils.Decrypt(req.Password, key)
		if err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		var usr user.User
		if err = svc.DB.First(&usr, "cube_id = ?", req.CubeID).Error; err != nil {
			exception.ErrUserNotFound.ResponseWithError(ctx, err)
			return
		}

		if req.InitPassword == "" {
			exception.ErrRegisterField.ResponseWithError(ctx, errors.New("该账号无法通过预设账号绑定"))
			return
		}

		if usr.InitPassword != req.InitPassword {
			exception.ErrRegisterField.ResponseWithError(ctx, errors.New("初始化密码不正确"))
			return
		}

		usr.InitPassword = ""
		usr.LoginID = req.LoginID
		usr.Password = password
		usr.HistoryPassword = password
		usr.QQ = req.QQ
		usr.Email = req.Email
		usr.WcaID = req.WcaID
		usr.Phone = req.Phone
		usr.Hash = string(utils.GenerateRandomKey(time.Now().UnixNano()))
		usr.ActivationTime = time.Now()

		if err = svc.DB.Save(&usr).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
