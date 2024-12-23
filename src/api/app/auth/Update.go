package auth

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type UpdateDetailReq struct {
	Name      string `json:"Name"`
	EnName    string `json:"EnName"`
	WcaID     string `json:"WcaID"`
	QQ        string `json:"QQ"`
	Sex       int    `json:"Sex"`
	Birthdate string `json:"Birthdate"`
	Sign      string `json:"Sign"`
}

func UpdateDetail(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}
		var req UpdateDetailReq
		if err = ctx.Bind(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		// todo 添加事务，可回滚

		var updateName = req.Name != user.Name

		if updateName && user.LastUpdateNameTime != nil && time.Since(*user.LastUpdateNameTime) < time.Minute*30 {
			exception.ErrAuthField.ResponseWithError(ctx, "30分钟内无法重复改名，请等待耐心等待")
			return
		}

		if updateName {
			err1 := svc.DB.Model(&result.Results{}).Where("cube_id = ?", user.CubeID).Update("person_name", req.Name).Error
			err2 := svc.DB.Model(&competition.Registration{}).Where("user_id = ?", user.ID).Update("user_name", req.Name).Error
			user.LastUpdateNameTime = utils.PtrNow()
			fmt.Println(err1, err2)
		}

		user.Name = req.Name
		user.EnName = req.EnName
		user.WcaID = req.WcaID
		user.QQ = req.QQ
		user.Sign = req.Sign
		user.Sex = req.Sex

		if req.Birthdate != "" {
			b, _ := time.Parse("2006-01-02", req.Birthdate)
			user.Birthdate = utils.PtrTime(b)
		}
		svc.DB.Save(&user)

	}
}
