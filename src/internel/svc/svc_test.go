package svc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewAPISvcWithoutTrainerFiles(t *testing.T) {
	config := filepath.Join(t.TempDir(), "server.yaml")
	if err := os.WriteFile(config, []byte("global:\n  algTrainersPath: ./build/Alg-Trainers\n  db:\n    driver: sqlite\n    dsn: catalog.db\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(t.TempDir())
	service, err := NewAPISvc(config, false, false, true)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := service.DB.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var tables []string
	if err := service.DB.Raw("SELECT name FROM sqlite_master WHERE type = 'table' AND name GLOB 'formula_*'").Scan(&tables).Error; err != nil {
		t.Fatal(err)
	}
	if len(tables) != 0 {
		t.Fatalf("shared service initialization created formula tables: %v", tables)
	}
}
