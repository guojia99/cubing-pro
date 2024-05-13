package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
)

const CubeProHeaderKey = "x-cubing-pro"

func CheckHeaderMiddleware(ctx *gin.Context) {
	cbH := ctx.GetHeader(CubeProHeaderKey)

	if cbH == "" {
		exception.ErrUnavailableService.ResponseWithError(ctx, "无操作权限")
		return
	}

	ctx.Next()
}
