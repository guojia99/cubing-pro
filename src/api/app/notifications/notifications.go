package notifications

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/formulaforum"
	"github.com/guojia99/cubing-pro/src/api/app/siteannouncement"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	model "github.com/guojia99/cubing-pro/src/internel/database/model/notification"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	formulaReply     = "formula-reply"
	siteAnnouncement = "site-announcement"
)

type notification struct {
	ID         string               `json:"id"`
	Type       string               `json:"type"`
	CreatedAt  time.Time            `json:"createdAt"`
	ReadAt     *time.Time           `json:"readAt"`
	Title      string               `json:"title,omitempty"`
	Body       string               `json:"body"`
	AuthorName string               `json:"authorName,omitempty"`
	CommentID  uint                 `json:"commentId,omitempty"`
	Target     *formulaforum.Target `json:"target,omitempty"`
}

type listRequest struct {
	Page int    `form:"page,default=1" binding:"min=1,max=100000000"`
	Size int    `form:"size,default=20" binding:"min=1,max=100"`
	Type string `form:"type,default=all" binding:"oneof=all formula-reply site-announcement"`
}

type page struct {
	Items  []notification `json:"items"`
	Total  int64          `json:"total"`
	Page   int            `json:"page"`
	Unread int64          `json:"unread"`
}

func List(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Cache-Control", "no-store")
		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}
		var req listRequest
		if err := ctx.ShouldBindQuery(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		var result page
		err = db.WithContext(ctx.Request.Context()).Transaction(func(tx *gorm.DB) error {
			var err error
			result, err = list(tx, user.ID, req)
			return err
		})
		if success(ctx, err) {
			exception.ResponseOK(ctx, result)
		}
	}
}

func list(db *gorm.DB, userID uint, req listRequest) (page, error) {
	result := page{Items: []notification{}, Page: req.Page}
	active, err := siteannouncement.Active(db)
	if err != nil {
		return result, err
	}
	reads := []model.SiteAnnouncementRead{}
	if err := db.Where("user_id = ?", userID).Find(&reads).Error; err != nil {
		return result, err
	}
	readAt := make(map[string]time.Time, len(reads))
	for _, read := range reads {
		readAt[read.AnnouncementID] = read.ReadAt
	}
	announcements := make([]notification, 0, len(active))
	for _, entry := range active {
		item := notification{ID: siteAnnouncement + ":" + entry.ID, Type: siteAnnouncement,
			CreatedAt: entry.PublishedAt, Title: entry.Title, Body: entry.Text}
		if read, ok := readAt[entry.ID]; ok {
			item.ReadAt = &read
		} else {
			result.Unread++
		}
		announcements = append(announcements, item)
	}
	var stats struct {
		Total  int64
		Unread int64
	}
	q := formulaforum.ReplyNotificationQuery(db, userID)
	if err := q.Session(&gorm.Session{}).Select("COUNT(*) AS total, COUNT(CASE WHEN n.read_at IS NULL THEN 1 END) AS unread").Scan(&stats).Error; err != nil {
		return result, err
	}
	result.Unread += stats.Unread
	if req.Type != formulaReply {
		result.Total += int64(len(announcements))
	}
	if req.Type != siteAnnouncement {
		result.Total += stats.Total
	}
	if result.Total == 0 {
		result.Page = 1
		return result, nil
	}
	if maxPage := int((result.Total-1)/int64(req.Size)) + 1; result.Page > maxPage {
		result.Page = maxPage
	}
	start := (result.Page - 1) * req.Size
	if req.Type == formulaReply {
		announcements = nil
	}
	items := announcements
	replyOffset := 0
	if req.Type != siteAnnouncement {
		// At most len(announcements) entries can move the requested reply window.
		// Keep queries bounded even when the requested page is far into the inbox.
		replyOffset = max(0, start-len(announcements))
		replies, err := formulaforum.ReplyNotifications(q.Offset(replyOffset).Limit(req.Size + len(announcements)).
			Order("c.created_at DESC, CAST(n.id AS CHAR) DESC"))
		if err != nil {
			return result, err
		}
		for _, reply := range replies {
			target := reply.Target
			items = append(items, notification{ID: formulaReply + ":" + strconv.FormatUint(uint64(reply.ID), 10),
				Type: formulaReply, CreatedAt: reply.CreatedAt, ReadAt: reply.ReadAt, Body: reply.Body,
				AuthorName: reply.AuthorName, CommentID: reply.CommentID, Target: &target})
		}
	}
	result.Items = notificationWindow(items, start-replyOffset, req.Size)
	return result, nil
}

func notificationWindow(items []notification, start, size int) []notification {
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].ID > items[j].ID
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	if start >= len(items) {
		return []notification{}
	}
	return items[start:min(start+size, len(items))]
}

func MarkRead(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}
		kind, id, ok := strings.Cut(ctx.Param("id"), ":")
		query := db.WithContext(ctx.Request.Context())
		switch {
		case ok && kind == formulaReply:
			value, parseErr := strconv.ParseUint(id, 10, strconv.IntSize)
			if parseErr != nil || value == 0 || strconv.FormatUint(value, 10) != id {
				exception.ErrRequestBinding.ResponseWithError(ctx, "通知 ID 无效")
				return
			}
			err = formulaforum.MarkReplyRead(query, user.ID, uint(value))
		case ok && kind == siteAnnouncement:
			if strings.TrimSpace(id) == "" {
				exception.ErrRequestBinding.ResponseWithError(ctx, "通知 ID 无效")
				return
			}
			err = markAnnouncementRead(query, user.ID, id)
		default:
			exception.ErrRequestBinding.ResponseWithError(ctx, "通知 ID 无效")
			return
		}
		if success(ctx, err) {
			exception.ResponseOK(ctx, gin.H{})
		}
	}
}

func markAnnouncementRead(db *gorm.DB, userID uint, id string) error {
	active, err := siteannouncement.Active(db)
	if err != nil {
		return err
	}
	for _, entry := range active {
		if entry.ID == id {
			return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.SiteAnnouncementRead{
				UserID: userID, AnnouncementID: id, ReadAt: time.Now().UTC(),
			}).Error
		}
	}
	return gorm.ErrRecordNotFound
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
