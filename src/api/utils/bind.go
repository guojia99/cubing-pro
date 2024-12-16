package app_utils

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
)

func BindAll(ctx *gin.Context, req interface{}) (err error) {
	defer func() {
		if err := recover(); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err.(error))
			return
		}
	}()

	_ = ctx.BindUri(req)
	_ = ctx.BindQuery(req)
	_ = ctx.Bind(req)

	return
}
