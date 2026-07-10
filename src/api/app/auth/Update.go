package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	user2 "github.com/guojia99/cubing-pro/src/internel/database/model/user"
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
			// 检查数据库名字是否有重复的
			var findUser user2.User
			if err = svc.DB.Where("name = ?", req.Name).First(&findUser).Error; err == nil && findUser.ID != user.ID {
				exception.ErrDatabase.ResponseWithError(ctx, "名字被使用了")
				return
			}

			if err = user2.PropagateUserName(svc.DB, user.ID, user.CubeID, req.Name); err != nil {
				exception.ErrDatabase.ResponseWithError(ctx, err)
				return
			}
			user.LastUpdateNameTime = utils.PtrNow()
		}

		user.Name = req.Name
		user.EnName = req.EnName
		//user.WcaID = req.WcaID
		user.QQ = req.QQ
		user.Sign = req.Sign
		user.Sex = req.Sex

		if req.Birthdate != "" {
			b, _ := time.Parse("2006-01-02", req.Birthdate)
			user.Birthdate = utils.PtrTime(b)
		}
		svc.DB.Save(&user)
		ctx.JSON(200, gin.H{})
	}
}
