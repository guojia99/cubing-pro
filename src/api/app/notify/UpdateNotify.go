package notify

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type UpdateNotifyReq struct {
	NotifyId uint   `uri:"notifyId"`
	Type     string `json:"type"`                             // 通知类型
	Title    string `json:"title" binding:"required,max=100"` // 标题
	Short    string `json:"short" binding:"max=50"`           // 简短
	Top      bool   `json:"top"`                              // 是否置顶
	Fixed    bool   `json:"fixed"`                            // 是否侧边
	Content  string `json:"content" binding:"required"`       // markdown
	Remark   string `json:"remark"`                           // 备注
}

func UpdateNotify(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req UpdateNotifyReq

		if err := ctx.ShouldBind(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		if err := ctx.BindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		var old post.Notification
		if err := svc.DB.First(&old, "id = ?", req.NotifyId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		old.Type = req.Type
		old.Title = req.Title
		old.Short = req.Short
		old.Short = req.Short
		old.Fixed = req.Fixed
		old.Content = req.Content
		old.Remark = req.Remark

		if err := svc.DB.Save(&old).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, nil)
	}
}
