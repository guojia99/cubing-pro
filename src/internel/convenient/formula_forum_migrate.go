package convenient

import (
	"fmt"

	"github.com/guojia99/cubing-pro/src/internel/database/model/formulaforum"

	"gorm.io/gorm"
)

const legacyFormulaDiscussionCheck = "chk_formula_discussion_target"

func migrateFormulaForum(db *gorm.DB) error {
	tables := []any{&formulaforum.FormulaDiscussion{}, &formulaforum.FormulaComment{}, &formulaforum.FormulaReplyNotification{}}
	if db.Migrator().HasConstraint(&formulaforum.FormulaDiscussion{}, legacyFormulaDiscussionCheck) {
		if !db.Migrator().HasColumn(&formulaforum.FormulaDiscussion{}, "SetID") {
			if err := db.Exec("ALTER TABLE formula_discussions ADD COLUMN set_id varchar(32)").Error; err != nil {
				return err
			}
		}
		if err := db.Exec(`ALTER TABLE formula_discussions DROP CHECK ` + legacyFormulaDiscussionCheck + `,
			ADD CONSTRAINT chk_formula_discussion_scope CHECK ((set_id IS NOT NULL) + (group_id IS NOT NULL) + (solution_id IS NOT NULL) = 1) ENFORCED`).Error; err != nil {
			return err
		}
	}
	if err := db.AutoMigrate(tables...); err != nil {
		return err
	}
	var enforced int64
	err := db.Raw(`SELECT COUNT(*) FROM information_schema.table_constraints WHERE table_schema = DATABASE()
		AND table_name = 'formula_discussions' AND constraint_type = 'CHECK'
		AND constraint_name = 'chk_formula_discussion_scope' AND enforced = 'YES'`).Scan(&enforced).Error
	if err != nil {
		return fmt.Errorf("checking formula discussion constraint (requires MySQL 8.0.16+): %w", err)
	}
	if enforced != 1 {
		return fmt.Errorf("formula discussion scope CHECK is not enforced")
	}
	return nil
}
