package formulaforum

import (
	"fmt"

	"gorm.io/gorm"
)

const legacyCheck = "chk_formula_discussion_target"

func Migrate(db *gorm.DB) error {
	tables := []any{&FormulaDiscussion{}, &FormulaComment{}, &FormulaReplyNotification{}}
	if db.Dialector.Name() == "sqlite" {
		return db.Transaction(func(tx *gorm.DB) error {
			for _, table := range tables {
				if !tx.Migrator().HasTable(table) {
					if err := tx.Migrator().CreateTable(table); err != nil {
						return err
					}
				}
			}
			if tx.Migrator().HasConstraint(&FormulaDiscussion{}, legacyCheck) {
				return upgradeSQLite(tx)
			}
			return nil
		})
	}
	if db.Dialector.Name() == "mysql" && db.Migrator().HasConstraint(&FormulaDiscussion{}, legacyCheck) {
		if !db.Migrator().HasColumn(&FormulaDiscussion{}, "SetID") {
			if err := db.Exec("ALTER TABLE formula_discussions ADD COLUMN set_id varchar(32)").Error; err != nil {
				return err
			}
		}
		if err := db.Exec(`ALTER TABLE formula_discussions DROP CHECK ` + legacyCheck + `,
			ADD CONSTRAINT chk_formula_discussion_scope CHECK ((set_id IS NOT NULL) + (group_id IS NOT NULL) + (solution_id IS NOT NULL) = 1) ENFORCED`).Error; err != nil {
			return err
		}
	}
	if err := db.AutoMigrate(tables...); err != nil {
		return err
	}
	if db.Dialector.Name() == "mysql" {
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
	}
	return nil
}

func upgradeSQLite(tx *gorm.DB) error {
	var objects []string
	err := tx.Raw(`SELECT sql FROM sqlite_master WHERE tbl_name = 'formula_discussions'
		AND type IN ('index', 'trigger') AND sql IS NOT NULL`).Scan(&objects).Error
	if err != nil {
		return err
	}
	var sequence int64
	if err := tx.Raw(`SELECT seq FROM sqlite_sequence WHERE name = 'formula_discussions'`).Scan(&sequence).Error; err != nil {
		return err
	}
	if err := tx.Exec(`ALTER TABLE formula_discussions ADD COLUMN set_id varchar(32)
		CONSTRAINT chk_formula_discussion_scope CHECK ((set_id IS NOT NULL) + (group_id IS NOT NULL) + (solution_id IS NOT NULL) = 1)`).Error; err != nil {
		return err
	}
	if err := tx.Migrator().DropConstraint(&FormulaDiscussion{}, legacyCheck); err != nil {
		return err
	}
	for _, ddl := range objects {
		if err := tx.Exec(ddl).Error; err != nil {
			return err
		}
	}
	if err := tx.Exec(`UPDATE sqlite_sequence SET seq = ? WHERE name = 'formula_discussions' AND seq < ?`, sequence, sequence).Error; err != nil {
		return err
	}
	return tx.Migrator().CreateIndex(&FormulaDiscussion{}, "SetID")
}
