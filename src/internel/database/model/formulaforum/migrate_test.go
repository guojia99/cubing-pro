package formulaforum

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type legacyDiscussion struct {
	ID         uint    `gorm:"primaryKey"`
	GroupID    *string `gorm:"type:varchar(32);uniqueIndex;check:chk_formula_discussion_target,(group_id IS NULL) <> (solution_id IS NULL)"`
	SolutionID *string `gorm:"type:varchar(32);uniqueIndex"`
}

func (legacyDiscussion) TableName() string { return "formula_discussions" }

func TestSQLiteDiscussionSchema(t *testing.T) {
	for _, scenario := range []string{"fresh", "upgrade"} {
		t.Run(scenario, func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "forum.db") + "?_foreign_keys=on"))
			if err != nil {
				t.Fatal(err)
			}
			if scenario == "upgrade" {
				testDiscussionUpgrade(t, db)
			} else if err := Migrate(db); err != nil {
				t.Fatal(err)
			}
			assertThreeScopes(t, db)
		})
	}
}

func TestMySQLDiscussionUpgrade(t *testing.T) {
	dsn := os.Getenv("FORMULA_FORUM_DISPOSABLE_MYSQL_DSN")
	if dsn == "" {
		t.Skip("requires an empty disposable MySQL database")
	}
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		t.Fatal(err)
	}
	tables, err := db.Migrator().GetTables()
	if err != nil || len(tables) != 0 {
		t.Fatalf("requires an empty disposable database: tables=%v, error=%v", tables, err)
	}
	testDiscussionUpgrade(t, db)
	assertThreeScopes(t, db)
}

func testDiscussionUpgrade(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.AutoMigrate(&legacyDiscussion{}, &FormulaComment{}, &FormulaReplyNotification{}); err != nil {
		t.Fatal(err)
	}
	group, solution := "0123456789abcdef", "fedcba9876543210"
	parent := uint(85)
	archived := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	for _, row := range []any{
		&legacyDiscussion{ID: 37, GroupID: &group}, &legacyDiscussion{ID: 63, SolutionID: &solution},
		&FormulaComment{ID: 71, DiscussionID: 37, AuthorID: 3, Body: "old group comment", DeletedAt: &archived},
		&FormulaComment{ID: parent, DiscussionID: 63, AuthorID: 2, Body: "parent"},
		&FormulaComment{ID: 91, DiscussionID: 63, AuthorID: 3, Body: "reply", ParentID: &parent},
		&FormulaReplyNotification{ID: 105, RecipientID: 2, CommentID: 91, ReadAt: &archived},
	} {
		if err := db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	if db.Dialector.Name() == "sqlite" {
		for _, sql := range []string{
			`CREATE INDEX saved_group_index ON formula_discussions(group_id) WHERE group_id IS NOT NULL`,
			`CREATE TRIGGER saved_trigger AFTER DELETE ON formula_discussions BEGIN SELECT 1; END`,
			`UPDATE sqlite_sequence SET seq = 199 WHERE name = 'formula_discussions'`,
		} {
			if err := db.Exec(sql).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	set := "aaaaaaaaaaaaaaaa"
	added := FormulaDiscussion{SetID: &set}
	if err := db.Create(&added).Error; err != nil {
		t.Fatal(err)
	}
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	if db.Dialector.Name() == "sqlite" {
		var count int
		if err := db.Raw(`SELECT COUNT(*) FROM sqlite_master WHERE name IN ('saved_group_index','saved_trigger')`).Scan(&count).Error; err != nil || count != 2 || added.ID != 200 {
			t.Fatalf("preserved objects=%d, next id=%d, error=%v", count, added.ID, err)
		}
	}
	for _, target := range []FormulaDiscussion{{ID: 37, GroupID: &group}, {ID: 63, SolutionID: &solution}, added} {
		var count int64
		if err := db.Model(&FormulaDiscussion{}).Where(&target).Count(&count).Error; err != nil || count != 1 {
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
