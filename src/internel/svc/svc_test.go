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
	var classes, timeout int
	if err := service.DB.Raw("SELECT COUNT(*) FROM formula_classes").Row().Scan(&classes); err != nil || classes != 0 {
		t.Fatalf("expected empty catalog schema: classes=%d, err=%v", classes, err)
	}
	if err := service.DB.Raw("PRAGMA busy_timeout").Row().Scan(&timeout); err != nil || timeout != 5000 {
		t.Fatalf("busy_timeout = %d, err=%v", timeout, err)
	}
}
