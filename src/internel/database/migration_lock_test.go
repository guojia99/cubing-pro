package database

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/mattn/go-sqlite3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestMySQLMigrationLock(t *testing.T) {
	migrationErr := errors.New("migration failed")
	for _, tc := range []struct {
		name       string
		acquired   any
		released   int
		migrateErr error
		wantErr    string
	}{
		{"success", 1, 1, nil, ""},
		{"timeout", 0, 1, nil, "acquire migration lock"},
		{"NULL result", nil, 1, nil, "acquire migration lock"},
		{"failed migration", 1, 1, migrationErr, "migration failed"},
		{"failed release", 1, 0, nil, "release migration lock"},
		{"both failures", 1, 0, migrationErr, "release migration lock"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool, err := sql.Open("sqlite3", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { pool.Close() })
			conn, err := pool.Conn(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			// Only this physical session has the lock functions and the temporary table.
			var calls []string
			err = conn.Raw(func(raw any) error {
				driver := raw.(*sqlite3.SQLiteConn)
				if err := driver.RegisterFunc("GET_LOCK", func(name string, timeout int) any {
					calls = append(calls, fmt.Sprintf("acquire %s %d", name, timeout))
					return tc.acquired
				}, false); err != nil {
					return err
				}
				return driver.RegisterFunc("RELEASE_LOCK", func(name string) int {
					calls = append(calls, "release "+name)
					return tc.released
				}, false)
			})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := conn.ExecContext(t.Context(), "CREATE TEMP TABLE session_marker (value INTEGER)"); err != nil {
				t.Fatal(err)
			}
			if err := conn.Close(); err != nil {
				t.Fatal(err)
			}
			db, err := gorm.Open(mysql.New(mysql.Config{Conn: pool, SkipInitializeWithVersion: true}), &gorm.Config{})
			if err != nil {
				t.Fatal(err)
			}
			err = WithMigrationLock(db, "catalog", func(session *gorm.DB) error {
				calls = append(calls, "migrate")
				return session.Transaction(func(tx *gorm.DB) error {
					if err := tx.Exec("INSERT INTO session_marker VALUES (1)").Error; err != nil {
						return err
					}
					return tc.migrateErr
				})
			})
			wantCalls := []string{"acquire catalog 120"}
			if tc.acquired == 1 {
				wantCalls = append(wantCalls, "migrate", "release catalog")
			}
			if !slices.Equal(calls, wantCalls) {
				t.Fatalf("calls = %v, want %v", calls, wantCalls)
			}
			if tc.wantErr == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.wantErr) || (tc.migrateErr != nil && !errors.Is(err, tc.migrateErr)) {
				t.Fatalf("error = %v, want %q and migration error %v", err, tc.wantErr, tc.migrateErr)
			}
		})
	}
}
