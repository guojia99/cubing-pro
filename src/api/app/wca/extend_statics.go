package wca

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/wca/types"
)

type ResultProportionEstimationRequest struct {
	EstimationType types.ResultProportionEstimationType `json:"estimationType" form:"estimationType"`
	WrN            int                                  `form:"wrN"`
}

func ResultProportionEstimation(svc *svc.Svc) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var req ResultProportionEstimationRequest
		if err := ctx.ShouldBindQuery(&req); err != nil {
			return
		}
		if req.WrN == 0 {
			req.WrN = 100
		}

		key := fmt.Sprintf("resultProportionEstimation-%s-%d", req.EstimationType, req.WrN)
		if out, ok := svc.Cache.Get(key); ok {
			exception.ResponseOK(ctx, out)
			return
		}

		out, err := svc.Wca.ResultProportionEstimation(req.EstimationType, req.WrN)
		if err != nil {
			exception.ErrGetData.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, out)
	}
}
