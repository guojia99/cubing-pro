package convenient

import (
	"testing"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/formulaforum"
	"github.com/guojia99/cubing-pro/src/internel/testutil"

	"gorm.io/gorm"
)

type legacyDiscussion struct {
	ID         uint    `gorm:"primaryKey"`
	GroupID    *string `gorm:"type:varchar(32);uniqueIndex;check:chk_formula_discussion_target,(group_id IS NULL) <> (solution_id IS NULL)"`
	SolutionID *string `gorm:"type:varchar(32);uniqueIndex"`
}

func (legacyDiscussion) TableName() string { return "formula_discussions" }

func TestMySQLDiscussionSchema(t *testing.T) {
	for _, scenario := range []string{"fresh", "upgrade"} {
		t.Run(scenario, func(t *testing.T) {
			db := testutil.OpenMySQL(t)
			if scenario == "upgrade" {
				testDiscussionUpgrade(t, db)
			} else if err := migrateFormulaForum(db); err != nil {
				t.Fatal(err)
			}
			assertThreeScopes(t, db)
		})
	}
}

func testDiscussionUpgrade(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.AutoMigrate(&legacyDiscussion{}, &formulaforum.FormulaComment{}, &formulaforum.FormulaReplyNotification{}); err != nil {
		t.Fatal(err)
	}
	group, solution := "0123456789abcdef", "fedcba9876543210"
	parent := uint(85)
	archived := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	for _, row := range []any{
		&legacyDiscussion{ID: 37, GroupID: &group}, &legacyDiscussion{ID: 63, SolutionID: &solution},
		&formulaforum.FormulaComment{ID: 71, DiscussionID: 37, AuthorID: 3, Body: "old group comment", DeletedAt: &archived},
		&formulaforum.FormulaComment{ID: parent, DiscussionID: 63, AuthorID: 2, Body: "parent"},
		&formulaforum.FormulaComment{ID: 91, DiscussionID: 63, AuthorID: 3, Body: "reply", ParentID: &parent},
		&formulaforum.FormulaReplyNotification{ID: 105, RecipientID: 2, CommentID: 91, ReadAt: &archived},
	} {
		if err := db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := migrateFormulaForum(db); err != nil {
		t.Fatal(err)
	}
	set := "aaaaaaaaaaaaaaaa"
	added := formulaforum.FormulaDiscussion{SetID: &set}
	if err := db.Create(&added).Error; err != nil {
		t.Fatal(err)
	}
	if err := migrateFormulaForum(db); err != nil {
		t.Fatal(err)
	}
	for _, target := range []formulaforum.FormulaDiscussion{{ID: 37, GroupID: &group}, {ID: 63, SolutionID: &solution}, added} {
		var count int64
		if err := db.Model(&formulaforum.FormulaDiscussion{}).Where(&target).Count(&count).Error; err != nil || count != 1 {
			t.Fatalf("target %d: count=%d, error=%v", target.ID, count, err)
		}
	}
	var count int
	if err := db.Raw(`SELECT COUNT(*) FROM formula_comments WHERE id = 71 AND discussion_id = 37
		AND body = 'old group comment' AND deleted_at = ?`, archived).Scan(&count).Error; err != nil || count != 1 {
		t.Fatalf("legacy group comment: count=%d, error=%v", count, err)
	}
	err := db.Raw(`SELECT COUNT(*) FROM formula_reply_notifications n JOIN formula_comments c ON c.id = n.comment_id
		JOIN formula_comments p ON p.id = c.parent_id WHERE n.id = 105 AND n.recipient_id = 2 AND c.id = 91 AND c.body = 'reply'
		AND c.discussion_id = 63 AND p.id = 85 AND p.discussion_id = 63 AND p.body = 'parent' AND n.read_at = ?`, archived).Scan(&count).Error
	if err != nil || count != 1 {
		t.Fatalf("legacy reply and notification: count=%d, error=%v", count, err)
	}
}

func assertThreeScopes(t *testing.T, db *gorm.DB) {
	t.Helper()
	const insert = `INSERT INTO formula_discussions(set_id,group_id,solution_id) VALUES(?,?,?)`
	const id = "bbbbbbbbbbbbbbbb"
	for _, values := range [][]any{{nil, nil, nil}, {id, id, nil}, {id, nil, id}, {nil, id, id}, {id, id, id}} {
		if err := db.Exec(insert, values...).Error; err == nil {
			t.Fatalf("accepted invalid discussion: %v", values)
		}
	}
	for _, values := range [][]any{{id, nil, nil}, {nil, id, nil}, {nil, nil, id}} {
		if err := db.Exec(insert, values...).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(insert, values...).Error; err == nil {
			t.Fatalf("accepted duplicate discussion: %v", values)
		}
	}
}
