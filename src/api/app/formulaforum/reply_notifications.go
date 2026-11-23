package formulaforum

import (
	"time"

	model "github.com/guojia99/cubing-pro/src/internel/database/model/formulaforum"
	"gorm.io/gorm"
)

type ReplyNotification struct {
	ID         uint       `json:"id"`
	CommentID  uint       `json:"commentId"`
	Target     Target     `json:"target" gorm:"embedded;embeddedPrefix:target_"`
	AuthorName string     `json:"authorName"`
	Body       string     `json:"body"`
	CreatedAt  time.Time  `json:"createdAt"`
	ReadAt     *time.Time `json:"readAt"`
}

func ReplyNotificationQuery(db *gorm.DB, recipient uint) *gorm.DB {
	return withCatalogOptional(db.Table("formula_reply_notifications n").
		Joins("JOIN formula_comments c ON c.id = n.comment_id AND c.deleted_at IS NULL").
		Joins("JOIN formula_discussions d ON d.id = c.discussion_id").Where("n.recipient_id = ?", recipient))
}

func ReplyNotifications(q *gorm.DB) ([]ReplyNotification, error) {
	items := []ReplyNotification{}
	err := q.Joins("LEFT JOIN users a ON a.id = c.author_id").
		Select(optionalTargetColumns + ", n.id, n.comment_id, n.read_at, c.created_at, c.body, COALESCE(a.name,'') AS author_name").
		Scan(&items).Error
	return items, err
}

func MarkReplyRead(db *gorm.DB, recipient, id uint) error {
	var n model.FormulaReplyNotification
	if err := db.Where("id = ? AND recipient_id = ?", id, recipient).Take(&n).Error; err != nil {
		return err
	}
	if n.ReadAt == nil {
		return db.Model(&n).Where("read_at IS NULL").Update("read_at", time.Now()).Error
	}
	return nil
}
