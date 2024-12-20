package auth

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
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
		user.Name = req.Name
		user.EnName = req.EnName
		user.WcaID = req.WcaID
		user.QQ = req.QQ
		user.Sign = req.Sign
		user.Sex = req.Sex
		fmt.Println(req.Sex)

		if req.Birthdate != "" {
			b, _ := time.Parse("2006-01-02", req.Birthdate)
			user.Birthdate = utils.PtrTime(b)
		}
		svc.DB.Save(&user)
	}
}
