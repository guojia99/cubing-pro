package other_link

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func GetOtherLinks(svc *svc.Svc) gin.HandlerFunc {
	return func(c *gin.Context) {
		var out OtherLinks
		if err := system.GetKeyJSONValue(svc.DB, otherLinkKey, &out); err != nil {
			c.JSON(200, out)
			return
		}
		c.JSON(200, out)
	}
}
