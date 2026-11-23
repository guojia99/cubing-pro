package formulaforum

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	model "github.com/guojia99/cubing-pro/src/internel/database/model/formulaforum"
	"gorm.io/gorm"
)

func MyDiscussions(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}
		var req paging
		if !bound(ctx, ctx.ShouldBindQuery(&req)) {
			return
		}
		query := db.WithContext(ctx.Request.Context())
		participated := query.Model(&model.FormulaComment{}).Select("discussion_id").
			Where("author_id = ? AND deleted_at IS NULL", user.ID)
		activity := query.Table("formula_comments activity_comment").Select(`discussion_id, COUNT(*) AS count,
			(SELECT latest.id FROM formula_comments latest WHERE latest.discussion_id = activity_comment.discussion_id
			AND latest.deleted_at IS NULL ORDER BY latest.created_at DESC, latest.id DESC LIMIT 1) AS last_id`).
			Where("deleted_at IS NULL AND discussion_id IN (?)", participated).Group("discussion_id")
		q := withCatalogOptional(query.Table("formula_discussions d").Joins("JOIN (?) activity ON activity.discussion_id = d.id", activity))
		var total int64
		if !success(ctx, q.Session(&gorm.Session{}).Count(&total).Error) {
			return
		}
		items := []struct {
			Target        Target    `json:"target" gorm:"embedded;embeddedPrefix:target_"`
			Count         int64     `json:"count"`
			LastActivity  time.Time `json:"lastActivity"`
			LastCommentID uint      `json:"lastCommentId"`
		}{}
		req.clamp(total)
		err = req.window(q).Joins("JOIN formula_comments c ON c.id = activity.last_id").
			Select(optionalTargetColumns + ", activity.count, c.created_at AS last_activity, c.id AS last_comment_id").
			Order("c.created_at DESC, c.id DESC").Scan(&items).Error
		if success(ctx, err) {
			exception.ResponseOK(ctx, gin.H{"items": items, "total": total, "page": req.Page})
		}
	}
}

func MyComments(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}
		var req paging
		if !bound(ctx, ctx.ShouldBindQuery(&req)) {
			return
		}
		q := withCatalogOptional(db.WithContext(ctx.Request.Context()).Table("formula_comments c").
			Joins("JOIN formula_discussions d ON d.id = c.discussion_id").
			Where("c.author_id = ? AND c.deleted_at IS NULL", user.ID))
		var total int64
		if !success(ctx, q.Session(&gorm.Session{}).Count(&total).Error) {
			return
		}
		items := []struct {
			Comment
			Target Target `json:"target" gorm:"embedded;embeddedPrefix:target_"`
		}{}
		req.clamp(total)
		err = req.window(withAuthors(q)).Select(commentColumns + ", " + optionalTargetColumns).
			Order("c.created_at DESC, c.id DESC").Scan(&items).Error
		if success(ctx, err) {
			exception.ResponseOK(ctx, gin.H{"items": items, "total": total, "page": req.Page})
		}
	}
}
