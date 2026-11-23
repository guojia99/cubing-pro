package api

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/guojia99/cubing-pro/src/internel/svc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestInitFormulaService(t *testing.T) {
	t.Run("empty path migrates schema only", func(t *testing.T) {
		t.Chdir(t.TempDir())
		service := openFormulaTestService(t, filepath.Join(t.TempDir(), "catalog.db"), "")
		if err := initFormulaService(service); err != nil {
			t.Fatal(err)
		}
		for _, table := range []string{
			"formula_catalog_imports", "formula_classes", "formula_sets", "formula_groups", "formula_cases", "formula_solutions",
			"formula_discussions", "formula_comments", "formula_reply_notifications",
		} {
			if !service.DB.Migrator().HasTable(table) {
				t.Errorf("missing table %s", table)
			}
		}
		assertFormulaCount(t, service.DB.Table("formula_classes"), 0)
		assertFormulaCount(t, service.DB.Table("formula_catalog_imports").Where("complete = ?", true), 0)
	})

	t.Run("configured path imports the catalog", func(t *testing.T) {
		root := t.TempDir()
		trainers := filepath.Join(root, "Alg-Trainers")
		library := filepath.Join(root, "Algs")
		if err := os.Mkdir(trainers, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.CopyFS(filepath.Join(library, "oll"), os.DirFS("../../build/Algs/algs/3x3/2-look-oll")); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(library, "catalog.json"), []byte(`[{"cube":"333","name":"2 Look OLL","path":"oll","cases":2}]`), 0600); err != nil {
			t.Fatal(err)
		}
		dbFile := filepath.Join(t.TempDir(), "catalog.db")
		service := openFormulaTestService(t, dbFile, trainers)
		if err := initFormulaService(service); err != nil {
			t.Fatal(err)
		}
		assertFormulaCount(t, service.DB.Table("formula_catalog_imports").Where("complete = ?", true), 1)
		assertFormulaCount(t, service.DB.Table("formula_classes").Where("cube = ? AND name = ?", "333", "2 Look OLL"), 1)
		assertFormulaCount(t, service.DB.Table("formula_cases"), 2)

		t.Run("restart skips missing source and keeps database edits", func(t *testing.T) {
			if err := service.DB.Table("formula_classes").Where("cube = ? AND name = ?", "333", "2 Look OLL").Update("name", "Edited OLL").Error; err != nil {
				t.Fatal(err)
			}
			restarted := openFormulaTestService(t, dbFile, filepath.Join(root, "missing"))
			if err := initFormulaService(restarted); err != nil {
				t.Fatal(err)
			}
			assertFormulaCount(t, restarted.DB.Table("formula_classes"), 1)
			assertFormulaCount(t, restarted.DB.Table("formula_classes").Where("name = ?", "Edited OLL"), 1)
			assertFormulaCount(t, restarted.DB.Table("formula_cases"), 2)
		})
	})

	t.Run("missing path fails initial import", func(t *testing.T) {
		root := t.TempDir()
		service := openFormulaTestService(t, filepath.Join(root, "catalog.db"), filepath.Join(root, "missing"))
		if err := initFormulaService(service); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("got %v, want missing import source error", err)
		}
		assertFormulaCount(t, service.DB.Table("formula_catalog_imports").Where("complete = ?", true), 0)
	})
}

func openFormulaTestService(t *testing.T, dbFile, trainerPath string) *svc.Svc {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(dbFile))
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := pool.Close(); err != nil {
			t.Error(err)
		}
	})
	service := &svc.Svc{DB: db}
	service.Cfg.GlobalConfig.AlgTrainersPath = trainerPath
	return service
}

func assertFormulaCount(t *testing.T, query *gorm.DB, want int64) {
	t.Helper()
	var count int64
	if err := query.Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != want {
		t.Errorf("%s rows = %d, want %d", query.Statement.Table, count, want)
	}
}
