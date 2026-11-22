package formulaforum

import (
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database"
	"gorm.io/gorm"
)

type FormulaDiscussion struct {
	ID         uint    `gorm:"primaryKey"`
	GroupID    *string `gorm:"type:varchar(32);uniqueIndex;check:chk_formula_discussion_target,(group_id IS NULL) <> (solution_id IS NULL)"`
	SolutionID *string `gorm:"type:varchar(32);uniqueIndex"`
}

type FormulaComment struct {
	ID           uint   `gorm:"primaryKey"`
	DiscussionID uint   `gorm:"not null;index:idx_formula_thread"`
	AuthorID     uint   `gorm:"not null"`
	ParentID     *uint  `gorm:"index"`
	Body         string `gorm:"size:400;not null"`
	CreatedAt    time.Time
	DeletedAt    *time.Time
}

type FormulaReplyNotification struct {
	ID          uint `gorm:"primaryKey"`
	RecipientID uint `gorm:"not null;index"`
	CommentID   uint `gorm:"not null;uniqueIndex"`
	ReadAt      *time.Time
}

func Migrate(db *gorm.DB) error {
	return database.WithMigrationLock(db, "cubing_pro_forum_migrate", func(conn *gorm.DB) error {
		return conn.AutoMigrate(&FormulaDiscussion{}, &FormulaComment{}, &FormulaReplyNotification{})
	})
}
