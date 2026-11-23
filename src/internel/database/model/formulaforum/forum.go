package formulaforum

import "time"

type FormulaDiscussion struct {
	ID         uint    `gorm:"primaryKey"`
	SetID      *string `gorm:"type:varchar(32);uniqueIndex;check:chk_formula_discussion_scope,(set_id IS NOT NULL) + (group_id IS NOT NULL) + (solution_id IS NOT NULL) = 1"`
	GroupID    *string `gorm:"type:varchar(32);uniqueIndex"`
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
