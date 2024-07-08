package result

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type PlayerSorReq struct {
	PlayerID uint     `uri:"playerId"`
	Events   []string `json:"Events"`
}

func PlayerSor(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req PlayerSorReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		sor, err := svc.Cov.KinChSorWithPlayer(req.PlayerID, req.Events)
		if err != nil {
			exception.ErrGetData.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, sor)
	}
}
