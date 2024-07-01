package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func Current(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}

		// 这里需要实时数据
		if err = svc.DB.Omit(user.AuthOmits()...).First(&user, "id = ?", user.ID).Error; err != nil {
			exception.ErrAuthField.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, user)
	}
}
