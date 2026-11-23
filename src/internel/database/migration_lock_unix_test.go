//go:build unix && ((!aix && !solaris) || illumos)

package database_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/guojia99/cubing-pro/src/internel/database"
	"gorm.io/gorm"
)

func TestSQLiteMigrationLockProcess(t *testing.T) {
	const probePath = "CUBING_TEST_SQLITE_FLOCK_PATH"
	if path := os.Getenv(probePath); path != "" {
		file, err := os.OpenFile(path, os.O_RDWR, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); !errors.Is(err, syscall.EWOULDBLOCK) {
			t.Fatalf("independent process lock attempt = %v, want EWOULDBLOCK", err)
		}
		return
	}
	path := filepath.Join(t.TempDir(), "catalog.db")
	err := database.WithMigrationLock(openSQLite(t, path), "catalog", func(*gorm.DB) error {
		child := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestSQLiteMigrationLockProcess$")
		child.Env = append(os.Environ(), probePath+"="+path+".migrate.lock")
		if out, err := child.CombinedOutput(); err != nil {
			return fmt.Errorf("child process: %w: %s", err, out)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
