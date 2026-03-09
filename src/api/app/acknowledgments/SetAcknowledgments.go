package acknowledgments

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func SetAcknowledgments(svc *svc.Svc) gin.HandlerFunc {
	return func(c *gin.Context) {

		var req Thanks
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		err := system.SetKeyJSONValue(svc.DB, thanksKey, req, "赞助列表")
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{})
	}
}
