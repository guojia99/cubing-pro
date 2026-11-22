package formulaforum

import (
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	model "github.com/guojia99/cubing-pro/src/internel/database/model/formulaforum"
	"gorm.io/gorm"
)

type paging struct {
	Page int `form:"page,default=1" binding:"min=1,max=100000000"`
	Size int `form:"size,default=20" binding:"min=1,max=100"`
}

func (p *paging) clamp(total int64) {
	if total <= 0 {
		p.Page = 1
		return
	}
	maxPage := int((total-1)/int64(p.Size)) + 1
	if p.Page > maxPage {
		p.Page = maxPage
	}
}

func (p paging) window(q *gorm.DB) *gorm.DB {
	return q.Offset((p.Page - 1) * p.Size).Limit(p.Size)
}

type resourceID struct {
	ID uint `uri:"id" binding:"required"`
}

func bound(ctx *gin.Context, err error) bool {
	if err != nil {
		exception.ErrRequestBinding.ResponseWithError(ctx, err)
		return false
	}
	return true
}

func success(ctx *gin.Context, err error) bool {
	switch {
	case err == nil:
		return true
	case errors.Is(err, gorm.ErrRecordNotFound):
		exception.ErrResourceNotFound.ResponseWithError(ctx, nil)
	default:
		exception.ErrDatabase.ResponseWithError(ctx, err)
	}
	return false
}

func Counts(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req struct {
			Cube  string `form:"cube" binding:"required"`
			Class string `form:"class" binding:"required"`
		}
		if !bound(ctx, ctx.ShouldBindQuery(&req)) {
			return
		}
		var rows []struct {
			GroupID    *string
			SolutionID *string
			Count      int64
		}
		q := db.WithContext(ctx.Request.Context()).Table("formula_discussions d").
			Joins("JOIN formula_comments c ON c.discussion_id = d.id AND c.deleted_at IS NULL")
		err := withCatalog(q).Select("d.group_id, d.solution_id, COUNT(*) AS count").
			Where("cl.cube = ? AND cl.name = ?", req.Cube, req.Class).Group("d.group_id, d.solution_id").Scan(&rows).Error
		if !success(ctx, err) {
			return
		}
		groups, solutions := map[string]int64{}, map[string]int64{}
		for _, row := range rows {
			if row.GroupID != nil {
				groups[*row.GroupID] = row.Count
			} else {
				solutions[*row.SolutionID] = row.Count
			}
		}
		exception.ResponseOK(ctx, gin.H{"groups": groups, "solutions": solutions})
	}
}

func Comments(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var key targetKey
		var req struct {
			paging
			CommentID uint `form:"commentId"`
		}
		if !bound(ctx, ctx.ShouldBindUri(&key)) || !bound(ctx, ctx.ShouldBindQuery(&req)) {
			return
		}
		query := db.WithContext(ctx.Request.Context())
		target, err := loadTarget(query, key)
		if !success(ctx, err) {
			return
		}
		thread := query.Model(&model.FormulaDiscussion{}).Select("id").Where(key.discussion())
		visible := query.Table("formula_comments c").Where("c.discussion_id = (?)", thread).
			Where("(c.deleted_at IS NULL OR EXISTS (SELECT 1 FROM formula_comments r WHERE r.parent_id = c.id AND r.deleted_at IS NULL))")
		var stats struct {
			Total int64
			Count int64
		}
		err = visible.Session(&gorm.Session{}).Select("COUNT(*) AS total, COUNT(CASE WHEN c.deleted_at IS NULL THEN 1 END) AS count").Scan(&stats).Error
		if !success(ctx, err) {
			return
		}

		req.clamp(stats.Total)
		if req.CommentID != 0 {
			var position, found int64
			err = visible.Session(&gorm.Session{}).Select("COUNT(*), COUNT(CASE WHEN c.id = ? THEN 1 END)", req.CommentID).
				Where("c.id <= ?", req.CommentID).Row().Scan(&position, &found)
			if !success(ctx, err) {
				return
			}
			if found != 0 {
				req.Page = int(position-1)/req.Size + 1
			}
		}

		items := []Comment{}
		err = withAuthors(req.window(visible)).Select(commentColumns).Order("c.id").Scan(&items).Error
		if !success(ctx, err) {
			return
		}
		exception.ResponseOK(ctx, gin.H{"target": target, "items": items, "total": stats.Total, "count": stats.Count, "page": req.Page})
	}
}

func Create(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}
		var key targetKey
		var req struct {
			Body     string `json:"body"`
			ParentID *uint  `json:"parentId" binding:"omitempty,gt=0"`
		}
		if !bound(ctx, ctx.ShouldBindUri(&key)) || !bound(ctx, ctx.ShouldBindJSON(&req)) {
			return
		}
		body := strings.TrimSpace(req.Body)
		if body == "" || utf8.RuneCountInString(body) > 400 {
			ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"code": "invalid_body", "message": "评论须为 1–400 个字"})
			return
		}
		for _, keyword := range blockedKeywords {
			if strings.Contains(body, keyword) {
				ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"code": "keyword_blocked", "message": "评论含有屏蔽词，请修改后再发表"})
				return
			}
		}
		id, err := createComment(db.WithContext(ctx.Request.Context()), key, user.ID, body, req.ParentID)
		if success(ctx, err) {
			exception.ResponseOK(ctx, gin.H{"id": id})
		}
	}
}

var blockedKeywords = []string{"刷单返利", "代开发票", "博彩开户", "色情引流", "傻逼", "操你妈"}

func Recent(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req paging
		if !bound(ctx, ctx.ShouldBindQuery(&req)) {
			return
		}
		query := db.WithContext(ctx.Request.Context())
		last := query.Model(&model.FormulaComment{}).Select("discussion_id, MAX(id) AS last_id, COUNT(*) AS count").
			Where("deleted_at IS NULL").Group("discussion_id")
		q := withCatalog(query.Table("formula_discussions d").Joins("JOIN (?) activity ON activity.discussion_id = d.id", last))
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
		err := req.window(q).Joins("JOIN formula_comments c ON c.id = activity.last_id").Select(targetColumns + ", activity.count, c.created_at AS last_activity, activity.last_id AS last_comment_id").
			Order("c.created_at DESC, c.id DESC").Scan(&items).Error
		if success(ctx, err) {
			exception.ResponseOK(ctx, gin.H{"items": items, "total": total, "page": req.Page})
		}
	}
}

func AdminComments(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req paging
		if !bound(ctx, ctx.ShouldBindQuery(&req)) {
			return
		}
		q := withCatalog(db.WithContext(ctx.Request.Context()).Table("formula_comments c").
			Joins("JOIN formula_discussions d ON d.id = c.discussion_id").Where("c.deleted_at IS NULL"))
		var total int64
		if !success(ctx, q.Session(&gorm.Session{}).Count(&total).Error) {
			return
		}
		items := []struct {
			Comment
			Target Target `json:"target" gorm:"embedded;embeddedPrefix:target_"`
		}{}
		req.clamp(total)
		err := req.window(withAuthors(q)).Select(commentColumns + ", " + targetColumns).Order("c.id DESC").Scan(&items).Error
		if success(ctx, err) {
			exception.ResponseOK(ctx, gin.H{"items": items, "total": total, "page": req.Page})
		}
	}
}

func Delete(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req resourceID
		if !bound(ctx, ctx.ShouldBindUri(&req)) {
			return
		}
		if success(ctx, removeComment(db.WithContext(ctx.Request.Context()), req.ID)) {
			exception.ResponseOK(ctx, gin.H{})
		}
	}
}

func Notifications(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}
		var req paging
		if !bound(ctx, ctx.ShouldBindQuery(&req)) {
			return
		}
		q := withCatalog(db.WithContext(ctx.Request.Context()).Table("formula_reply_notifications n").
			Joins("JOIN formula_comments c ON c.id = n.comment_id").Joins("JOIN formula_discussions d ON d.id = c.discussion_id").
			Where("n.recipient_id = ?", user.ID))
		var stats struct {
			Total  int64
			Unread int64
		}
		err = q.Session(&gorm.Session{}).Select("COUNT(*) AS total, COUNT(CASE WHEN n.read_at IS NULL THEN 1 END) AS unread").Scan(&stats).Error
		if !success(ctx, err) {
			return
		}
		items := []struct {
			ID         uint       `json:"id"`
			CommentID  uint       `json:"commentId"`
			Target     Target     `json:"target" gorm:"embedded;embeddedPrefix:target_"`
			AuthorName string     `json:"authorName"`
			Body       string     `json:"body"`
			CreatedAt  time.Time  `json:"createdAt"`
			ReadAt     *time.Time `json:"readAt"`
		}{}
		req.clamp(stats.Total)
		err = req.window(q).Joins("LEFT JOIN users a ON a.id = c.author_id").
			Select(targetColumns + ", n.id, n.comment_id, n.read_at, c.created_at, c.body, COALESCE(a.name,'') AS author_name").
			Order("n.id DESC").Scan(&items).Error
		if success(ctx, err) {
			exception.ResponseOK(ctx, gin.H{"items": items, "total": stats.Total, "unread": stats.Unread, "page": req.Page})
		}
	}
}

func MarkRead(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}
		var req resourceID
		if !bound(ctx, ctx.ShouldBindUri(&req)) {
			return
		}
		var n model.FormulaReplyNotification
		query := db.WithContext(ctx.Request.Context())
		err = query.Where("id = ? AND recipient_id = ?", req.ID, user.ID).Take(&n).Error
		if !success(ctx, err) {
			return
		}
		if n.ReadAt == nil {
			err = query.Model(&n).Where("read_at IS NULL").Update("read_at", time.Now()).Error
		}
		if success(ctx, err) {
			exception.ResponseOK(ctx, gin.H{})
		}
	}
}
