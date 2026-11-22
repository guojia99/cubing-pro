package database

import (
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// WithMigrationLock holds a MySQL session lock across DDL and any import transaction.
func WithMigrationLock(db *gorm.DB, name string, migrate func(*gorm.DB) error) error {
	if db.Dialector.Name() != "mysql" {
		return migrate(db)
	}
	return db.Connection(func(conn *gorm.DB) (err error) {
		// Start each query with a fresh statement on the pinned connection.
		conn = conn.Session(&gorm.Session{NewDB: true})
		if err := checkLockResult(conn.Raw("SELECT GET_LOCK(?, 120)", name).Row()); err != nil {
			return fmt.Errorf("acquire migration lock %q: %w", name, err)
		}
		defer func() {
			if releaseErr := checkLockResult(conn.Raw("SELECT RELEASE_LOCK(?)", name).Row()); releaseErr != nil {
				err = errors.Join(err, fmt.Errorf("release migration lock %q: %w", name, releaseErr))
			}
		}()
		return migrate(conn)
	})
}

func checkLockResult(row *sql.Row) error {
	var result int
	if err := row.Scan(&result); err != nil {
		return err
	}
	if result != 1 {
		return fmt.Errorf("expected 1, got %d", result)
	}
	return nil
}
