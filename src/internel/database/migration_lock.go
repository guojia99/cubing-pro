package database

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var sqliteMigrationLocks sync.Map // absolute database path -> *sync.Mutex

// WithMigrationLock serializes DDL and import transactions across service processes.
func WithMigrationLock(db *gorm.DB, name string, migrate func(*gorm.DB) error) error {
	if dialect, ok := db.Dialector.(*sqlite.Dialector); ok {
		path, err := sqliteDatabasePath(dialect.DSN)
		if err != nil {
			return fmt.Errorf("resolve SQLite migration lock path: %w", err)
		}
		if path == "" {
			return migrate(db)
		}
		entry, _ := sqliteMigrationLocks.LoadOrStore(path, new(sync.Mutex))
		mutex := entry.(*sync.Mutex)
		mutex.Lock()
		defer mutex.Unlock()
		file, err := os.OpenFile(path+".migrate.lock", os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			return err
		}
		return withSQLiteFileLock(file, func() error { return migrate(db) })
	}
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

func withSQLiteFileLock(file *os.File, migrate func() error) (err error) {
	// Close releases the OS lock. Never unlink the file: waiters may still hold its inode.
	defer func() { err = errors.Join(err, file.Close()) }()
	if err := lockSQLiteFile(file); err != nil {
		return &os.PathError{Op: "lock", Path: file.Name(), Err: err}
	}
	return migrate()
}

func sqliteDatabasePath(dsn string) (string, error) {
	path, query, _ := strings.Cut(dsn, "?")
	options, err := url.ParseQuery(query)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(path, "file:") {
		path = strings.TrimPrefix(path, "file:")
		if strings.HasPrefix(path, "//localhost/") {
			path = strings.TrimPrefix(path, "//localhost")
		} else if strings.HasPrefix(path, "///") {
			path = path[2:]
		}
		path, err = url.PathUnescape(path)
		if err != nil {
			return "", err
		}
		if strings.HasPrefix(path, "/") && filepath.VolumeName(path[1:]) != "" {
			path = path[1:]
		}
	}
	if path == "" || path == ":memory:" || options.Get("mode") == "memory" {
		return "", nil
	}
	return filepath.Abs(path)
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
