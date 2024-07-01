package notify

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type CreateNotifyReq struct {
	Title   string `json:"title" binding:"required,max=100"` // 标题
	Short   string `json:"short" binding:"max=50"`           // 简短
	Type    string `json:"type"`                             // 通知类型
	Top     bool   `json:"top"`                              // 是否置顶
	Fixed   bool   `json:"fixed"`                            // 是否侧边
	Content string `json:"content"`                          // markdown
	Remark  string `json:"remark"`                           // 备注
}

func CreateNotify(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}

		var req CreateNotifyReq
		if err = app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		notify := post.Notification{
			Title:          req.Title,
			Short:          req.Short,
			Type:           req.Type,
			Top:            req.Top,
			Fixed:          req.Fixed,
			Content:        req.Content,
			CreateBy:       user.Name,
			CreateByUserID: user.ID,
			Remark:         req.Remark,
		}

		if err = svc.DB.Create(&notify).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, nil)
	}
}
