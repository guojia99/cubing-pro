package comp

import (
	"github.com/gin-gonic/gin"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type RegistersReq struct {
	CompReq
}

func Registers(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req RegistersReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var regs []competition.Registration

		app_utils.GenerallyList(
			ctx, svc.DB, regs, app_utils.ListSearchParam[competition.Registration]{
				Model:   &competition.Registration{},
				MaxSize: 0,
				Query:   "comp_id = ? and status = ?",
				QueryCons: []interface{}{
					req.CompId, competition.RegisterStatusPass,
				},
				Select: []string{
					"user_id", "user_name", "events",
				},
			},
		)
	}
}
