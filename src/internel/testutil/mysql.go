package testutil

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"os"
	"testing"

	drivermysql "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// OpenMySQL creates a separate database for each test and drops it at cleanup.
// CUBING_TEST_MYSQL_DSN must point to a test server with CREATE/DROP DATABASE privileges.
func OpenMySQL(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("CUBING_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set CUBING_TEST_MYSQL_DSN to run MySQL integration tests")
	}
	cfg, err := drivermysql.ParseDSN(dsn)
	if err != nil {
		t.Fatalf("parse test MySQL DSN: %v", err)
	}
	cfg.DBName = ""
	cfg.ParseTime = true
	admin, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatalf("open test MySQL server: %v", err)
	}
	t.Cleanup(func() { _ = admin.Close() })
	var suffix [12]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		t.Fatal(err)
	}
	name := "cubing_test_" + hex.EncodeToString(suffix[:])
	if _, err := admin.Exec("CREATE DATABASE `" + name + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"); err != nil {
		t.Fatalf("create isolated MySQL test database: %v", err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec("DROP DATABASE `" + name + "`"); err != nil {
			t.Errorf("drop MySQL test database: %v", err)
		}
	})
	cfg.DBName = name
	db, err := gorm.Open(mysql.Open(cfg.FormatDSN()), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("open isolated MySQL test database: %v", err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	return db
}
