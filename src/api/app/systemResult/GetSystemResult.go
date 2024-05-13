package systemResults

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

const baseKey = "cubing-pro-"

const (
	systemTitleKey   = "cubing-pro-title"
	systemWelcomeKey = "cubing-pro-welcome"
	systemFooterKey  = "cubing-pro-footer"
	systemLogoKey    = "cubing-pro-logo"
)

func GetSystemResult(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var keyValue []system.KeyValue
		svc.DB.Find(&keyValue)
		exception.ResponseOK(ctx, keyValue)
	}
}
