package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type CompPlayersReq struct {
	CompReq
	//Apply bool `json:"Apply"`
}

func CompPlayers(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CompPlayersReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var regs []competition.Registration
		if err := svc.DB.Find(&regs, "comp_id = ?", req.CompId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, regs)
	}
}
