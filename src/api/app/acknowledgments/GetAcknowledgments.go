package acknowledgments

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func GetAcknowledgments(svc *svc.Svc) gin.HandlerFunc {
	return func(c *gin.Context) {
		var out Thanks
		if err := system.GetKeyJSONValue(svc.DB, thanksKey, &out); err != nil {
			c.JSON(200, out)
			return
		}
		c.JSON(200, out)
	}
}
