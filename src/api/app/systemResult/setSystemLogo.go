package systemResults

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func SetSystemLogo(svc *svc.Svc) gin.HandlerFunc {
	return setKeyValue(systemLogoKey, svc)
}
