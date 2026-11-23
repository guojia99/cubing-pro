package formulaforum

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/algs"
	model "github.com/guojia99/cubing-pro/src/internel/database/model/formulaforum"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestVirtualEGShelfHasIndependentDiscussion(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "eg.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	conn, err := db.DB()
	must(err)
	t.Cleanup(func() { _ = conn.Close() })
	must(algs.MigrateCatalog(db))
	must(model.Migrate(db))

	eg := targetKey{Kind: "class", ID: egShelfDiscussionID}
	if _, err := loadTarget(db, eg); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("virtual shelf exists without its catalog: %v", err)
	}

	const (
		eg1      = "1111111111111111"
		eg2      = "2222222222222222"
		eg1Set   = "3333333333333333"
		eg1Group = "4444444444444444"
		eg1Case  = "5555555555555555"
		eg1Sol   = "6666666666666666"
	)
	for _, row := range []struct {
		table  string
		values map[string]any
	}{
		{"formula_classes", map[string]any{"id": eg1, "cube": "222", "name": "EG1", "position": 0}},
		{"formula_classes", map[string]any{"id": eg2, "cube": "222", "name": "EG2", "position": 1}},
		{"formula_sets", map[string]any{"id": eg1Set, "class_id": eg1, "name": "EG1", "position": 0}},
		{"formula_groups", map[string]any{"id": eg1Group, "set_id": eg1Set, "name": "T", "position": 0}},
		{"formula_cases", map[string]any{"id": eg1Case, "group_id": eg1Group, "name": "T1", "position": 0}},
		{"formula_solutions", map[string]any{"id": eg1Sol, "case_id": eg1Case, "position": 0, "moves": "R U R'"}},
	} {
		must(db.Table(row.table).Create(row.values).Error)
	}

	target, err := loadTarget(db, eg)
	must(err)
	if want := (Target{Kind: "class", ID: egShelfDiscussionID, Cube: "222", ClassName: "EG"}); target != want {
		t.Fatalf("virtual class metadata %+v, want %+v", target, want)
	}

	engine := gin.New()
	engine.GET("/counts", Counts(db))
	type countsPayload struct {
		ClassID    string         `json:"classId"`
		ClassCount int            `json:"classCount"`
		Groups     map[string]int `json:"groups"`
		Solutions  map[string]int `json:"solutions"`
	}
	counts := func(label string) countsPayload {
		t.Helper()
		out := httptest.NewRecorder()
		engine.ServeHTTP(out, httptest.NewRequest(http.MethodGet, "/counts?cube=222&class="+label, nil))
		if out.Code != 200 {
			t.Fatalf("counts %s: %d %s", label, out.Code, out.Body.String())
		}
		var parsed struct{ Data countsPayload }
		must(json.Unmarshal(out.Body.Bytes(), &parsed))
		return parsed.Data
	}
	if got := counts("EG"); got.ClassID != egShelfDiscussionID || got.ClassCount != 0 {
		t.Fatalf("virtual empty counts missing the stable ID: %+v", got)
	}

	parent, err := createComment(db, eg, 1, "整套 EG", nil)
	must(err)
	reply, err := createComment(db, eg, 2, "整套 EG 追评", &parent)
	must(err)
	for _, k := range []targetKey{{Kind: "class", ID: eg1}, {Kind: "group", ID: eg1Group}, {Kind: "solution", ID: eg1Sol}} {
		if _, err := createComment(db, k, 3, "物理分类讨论", &parent); !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("reply crossed discussion scopes (%+v): %v", k, err)
		}
		_, err := createComment(db, k, 3, "物理分类讨论", nil)
		must(err)
	}
	if got := counts("EG"); got.ClassID != egShelfDiscussionID || got.ClassCount != 2 || len(got.Groups) != 0 || len(got.Solutions) != 0 {
		t.Fatalf("virtual counts mixed with physical discussions: %+v", got)
	}
	if got := counts("EG1"); got.ClassID != eg1 || got.ClassCount != 1 || got.Groups[eg1Group] != 1 || got.Solutions[eg1Sol] != 1 {
		t.Fatalf("physical EG1 counts mixed with virtual discussion: %+v", got)
	}
	var notice model.FormulaReplyNotification
	must(db.Where("comment_id = ?", reply).First(&notice).Error)
	if notice.RecipientID != 1 {
		t.Fatalf("class reply notification recipient: %d", notice.RecipientID)
	}
	must(removeComment(db, reply))
	if got := counts("EG"); got.ClassCount != 1 {
		t.Fatalf("deleted class reply counted: %+v", got)
	}
	if err := db.First(&notice, notice.ID).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("notification for deleted class reply still present: %v", err)
	}
}
