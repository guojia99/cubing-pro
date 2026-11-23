package api

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/guojia99/cubing-pro/src/internel/convenient"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/testutil"
	"gorm.io/gorm"
)

func TestNewAPIInitializesFormulaModules(t *testing.T) {
	t.Run("convenient migrates schema before API initialization", func(t *testing.T) {
		t.Chdir(t.TempDir())
		service := openFormulaTestService(t, nil, "")
		for _, table := range []string{
			"formula_catalog_imports", "formula_classes", "formula_sets", "formula_groups", "formula_cases", "formula_solutions",
			"formula_discussions", "formula_comments", "formula_reply_notifications",
		} {
			if !service.DB.Migrator().HasTable(table) {
				t.Errorf("missing table %s", table)
			}
		}
		if _, err := NewAPI(service); err != nil {
			t.Fatal(err)
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
		service := openFormulaTestService(t, nil, trainers)
		if _, err := NewAPI(service); err != nil {
			t.Fatal(err)
		}
		assertFormulaCount(t, service.DB.Table("formula_catalog_imports").Where("complete = ?", true), 1)
		assertFormulaCount(t, service.DB.Table("formula_classes").Where("cube = ? AND name = ?", "333", "2 Look OLL"), 1)
		assertFormulaCount(t, service.DB.Table("formula_cases"), 2)

		t.Run("restart skips missing source and keeps database edits", func(t *testing.T) {
			if err := service.DB.Table("formula_classes").Where("cube = ? AND name = ?", "333", "2 Look OLL").Update("name", "Edited OLL").Error; err != nil {
				t.Fatal(err)
			}
			restarted := openFormulaTestService(t, service.DB, filepath.Join(root, "missing"))
			if _, err := NewAPI(restarted); err != nil {
				t.Fatal(err)
			}
			assertFormulaCount(t, restarted.DB.Table("formula_classes"), 1)
			assertFormulaCount(t, restarted.DB.Table("formula_classes").Where("name = ?", "Edited OLL"), 1)
			assertFormulaCount(t, restarted.DB.Table("formula_cases"), 2)
		})
	})

	t.Run("missing path fails initial import", func(t *testing.T) {
		root := t.TempDir()
		service := openFormulaTestService(t, nil, filepath.Join(root, "missing"))
		if api, err := NewAPI(service); !errors.Is(err, os.ErrNotExist) || api != nil {
			t.Fatalf("got %v, want missing import source error", err)
		}
		assertFormulaCount(t, service.DB.Table("formula_catalog_imports").Where("complete = ?", true), 0)
	})
}

func openFormulaTestService(t *testing.T, db *gorm.DB, trainerPath string) *svc.Svc {
	t.Helper()
	if db == nil {
		db = testutil.OpenMySQL(t)
	}
	service := &svc.Svc{DB: db}
	service.Cfg.GlobalConfig.AlgTrainersPath = trainerPath
	service.Cov = convenient.NewConvenient(db, false, service.Cfg)
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
