package result

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type PlayerNemesisReq struct {
	PlayerID uint     `uri:"playerId"`
	Events   []string `json:"Events"`
}

func PlayerNemesis(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req PlayerNemesisReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		nemesis := svc.Cov.PlayerNemesisWithID(req.PlayerID, req.Events)
		exception.ResponseOK(ctx, nemesis)
	}
}
