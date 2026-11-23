package formulaforum

import (
	"time"

	model "github.com/guojia99/cubing-pro/src/internel/database/model/formulaforum"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Target struct {
	Kind        string `json:"kind"`
	ID          string `json:"id"`
	Cube        string `json:"cube"`
	ClassName   string `json:"className"`
	SetName     string `json:"setName"`
	GroupName   string `json:"groupName"`
	CaseName    string `json:"caseName"`
	Moves       string `json:"moves"`
	Unavailable bool   `json:"unavailable,omitempty"`
}

type targetKey struct {
	Kind string `uri:"kind" binding:"required,oneof=group solution"`
	ID   string `uri:"targetId" binding:"required,len=16,hexadecimal,lowercase"`
}

func (k targetKey) discussion() model.FormulaDiscussion {
	if k.Kind == "group" {
		return model.FormulaDiscussion{GroupID: &k.ID}
	}
	return model.FormulaDiscussion{SolutionID: &k.ID}
}

const targetColumns = `CASE WHEN d.solution_id IS NULL THEN 'group' ELSE 'solution' END AS target_kind,
	COALESCE(d.solution_id, d.group_id) AS target_id, COALESCE(cl.cube, '') AS target_cube, COALESCE(cl.name, '') AS target_class_name,
	COALESCE(st.name, '') AS target_set_name, COALESCE(g.name, '') AS target_group_name, COALESCE(ca.name, '') AS target_case_name,
	COALESCE(s.moves, '') AS target_moves`

const optionalTargetColumns = targetColumns + `, CASE WHEN g.id IS NULL OR
	(d.solution_id IS NOT NULL AND (s.id IS NULL OR LENGTH(s.moves) = 0)) THEN 1 ELSE 0 END AS target_unavailable`

func withCatalog(q *gorm.DB) *gorm.DB {
	return q.Joins("LEFT JOIN formula_solutions s ON s.id = d.solution_id AND LENGTH(s.moves) > 0").
		Joins("LEFT JOIN formula_cases ca ON ca.id = s.case_id").
		Joins("JOIN formula_groups g ON g.id = COALESCE(d.group_id, ca.group_id)").
		Joins("JOIN formula_sets st ON st.id = g.set_id").
		Joins("JOIN formula_classes cl ON cl.id = st.class_id")
}

func withCatalogOptional(q *gorm.DB) *gorm.DB {
	return q.Joins("LEFT JOIN formula_solutions s ON s.id = d.solution_id").
		Joins("LEFT JOIN formula_cases ca ON ca.id = s.case_id").
		Joins("LEFT JOIN formula_groups g ON g.id = COALESCE(d.group_id, ca.group_id)").
		Joins("LEFT JOIN formula_sets st ON st.id = g.set_id").
		Joins("LEFT JOIN formula_classes cl ON cl.id = st.class_id")
}

type targetRow struct {
	Target Target `json:"target" gorm:"embedded;embeddedPrefix:target_"`
}

func loadTarget(db *gorm.DB, key targetKey) (Target, error) {
	d := key.discussion()
	q := db.Table("(SELECT ? AS group_id, ? AS solution_id) AS d", d.GroupID, d.SolutionID)
	var row targetRow
	result := withCatalog(q).Select(targetColumns).Scan(&row)
	if result.Error == nil && result.RowsAffected == 0 {
		return Target{}, gorm.ErrRecordNotFound
	}
	return row.Target, result.Error
}

type Comment struct {
	ID          uint      `json:"id"`
	ParentID    *uint     `json:"parentId"`
	AuthorID    uint      `json:"authorId"`
	AuthorName  string    `json:"authorName"`
	ReplyToName string    `json:"replyToName"`
	Body        string    `json:"body"`
	Deleted     bool      `json:"deleted"`
	CreatedAt   time.Time `json:"createdAt"`
}

const commentColumns = `c.id, c.parent_id, c.created_at, c.deleted_at IS NOT NULL AS deleted,
	CASE WHEN c.deleted_at IS NULL THEN c.author_id ELSE 0 END AS author_id,
	CASE WHEN c.deleted_at IS NULL THEN COALESCE(a.name, '') ELSE '' END AS author_name,
	CASE WHEN c.deleted_at IS NULL AND parent.deleted_at IS NULL THEN COALESCE(pa.name, '') ELSE '' END AS reply_to_name,
	CASE WHEN c.deleted_at IS NULL THEN c.body ELSE '' END AS body`

func withAuthors(q *gorm.DB) *gorm.DB {
	return q.Joins("LEFT JOIN users a ON a.id = c.author_id").
		Joins("LEFT JOIN formula_comments parent ON parent.id = c.parent_id").
		Joins("LEFT JOIN users pa ON pa.id = parent.author_id")
}

func createComment(db *gorm.DB, key targetKey, author uint, body string, parentID *uint) (uint, error) {
	var comment model.FormulaComment
	err := db.Transaction(func(tx *gorm.DB) error {
		if _, err := loadTarget(tx, key); err != nil {
			return err
		}

		d := key.discussion()
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&d).Error; err != nil {
			return err
		}
		var discussion model.FormulaDiscussion
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(key.discussion()).Take(&discussion).Error; err != nil {
			return err
		}
		var parent model.FormulaComment
		if parentID != nil {
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("discussion_id = ? AND id = ? AND deleted_at IS NULL", discussion.ID, *parentID).Take(&parent).Error
			if err != nil {
				return err
			}
		}
		comment = model.FormulaComment{DiscussionID: discussion.ID, ParentID: parentID, AuthorID: author, Body: body}
		if err := tx.Create(&comment).Error; err != nil {
			return err
		}
		if parentID != nil && parent.AuthorID != author {
			return tx.Create(&model.FormulaReplyNotification{RecipientID: parent.AuthorID, CommentID: comment.ID}).Error
		}
		return nil
	})
	return comment.ID, err
}

func removeComment(db *gorm.DB, id uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.FormulaComment{}).Where("id = ? AND deleted_at IS NULL", id).Update("deleted_at", time.Now())
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Where("comment_id = ?", id).Delete(&model.FormulaReplyNotification{}).Error
	})
}
