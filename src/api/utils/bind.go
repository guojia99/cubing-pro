package app_utils

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
)

func BindAll(ctx *gin.Context, req interface{}) (err error) {
	defer func() {
		if err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
		}
	}()

	if err = ctx.ShouldBindUri(&req); err != nil {
		return err
	}
	if err = ctx.ShouldBindQuery(&req); err != nil {
		return err
	}
	if err = ctx.ShouldBind(&req); err != nil {
		return err
	}
	return nil
}
