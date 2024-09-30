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

	if err = ctx.BindUri(&req); err != nil {
		return err
	}
	if err = ctx.BindQuery(&req); err != nil {
		return err
	}
	if err = ctx.Bind(&req); err != nil {
		return err
	}
	return nil
}
