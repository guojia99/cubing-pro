package initer

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestInitFormulas(t *testing.T) {
	assets := t.TempDir()
	trainers := filepath.Join(assets, "Alg-Trainers")
	library := filepath.Join(assets, "Algs")
	if err := os.Mkdir(trainers, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.CopyFS(filepath.Join(library, "oll"), os.DirFS("../../build/Algs/algs/3x3/2-look-oll")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(library, "catalog.json"), []byte(`[{"cube":"333","name":"2 Look OLL","path":"oll","cases":2}]`), 0600); err != nil {
		t.Fatal(err)
	}
	cwd := t.TempDir()
	t.Chdir(cwd)
	for _, pathSource := range []string{"config", "flag"} {
		t.Run(pathSource, func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "catalog.db")))
			if err != nil {
				t.Fatal(err)
			}
			pool, err := db.DB()
			if err != nil {
				t.Fatal(err)
			}
			defer pool.Close()
			run := func(configPath string, args ...string) (string, error) {
				var service *svc.Svc
				root := &cobra.Command{Use: "cubing-pro", SilenceUsage: true, SilenceErrors: true, PersistentPreRun: func(*cobra.Command, []string) {
					service = &svc.Svc{DB: db}
					service.Cfg.GlobalConfig.AlgTrainersPath = configPath
				}}
				root.AddCommand(NewCmd(&service))
				root.SetArgs(append([]string{"init", "formulas"}, args...))
				var output bytes.Buffer
				root.SetOut(&output)
				err := root.Execute()
				return output.String(), err
			}
			var args []string
			configPath := trainers
			if pathSource == "flag" {
				if _, err := run(""); err == nil || !strings.Contains(err.Error(), "--path") || !strings.Contains(err.Error(), "algTrainersPath") {
					t.Fatalf("empty path error = %v", err)
				}
				if _, err := run(trainers, "--path", "missing"); err == nil || !strings.Contains(err.Error(), filepath.Join(cwd, "missing")) || !strings.Contains(err.Error(), filepath.Join(cwd, "Algs", "catalog.json")) {
					t.Fatalf("missing source error = %v", err)
				}
				configPath, args = "missing", []string{"-p", trainers}
			}
			if output, err := run(configPath, args...); err != nil || !strings.Contains(output, "1 个分类，2 个案例") {
				t.Fatalf("import: output=%q, err=%v", output, err)
			}
			if err := db.Exec("UPDATE formula_classes SET name = 'Edited'").Error; err != nil {
				t.Fatal(err)
			}
			if output, err := run(configPath, args...); err != nil || !strings.Contains(output, "1 个分类，2 个案例") {
				t.Fatalf("repeat import: output=%q, err=%v", output, err)
			}
			var name string
			if err := db.Raw("SELECT name FROM formula_classes").Row().Scan(&name); err != nil || name != "Edited" {
				t.Fatalf("repeat import did not retain database edit: name=%q, err=%v", name, err)
			}
		})
	}
}
