package posts

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func GetForums(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var forums []post.Forum
		svc.DB.Find(&forums)
		exception.ResponseOK(ctx, forums)
	}
}
