package public

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/algs"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/patrickmn/go-cache"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSyncedAlgorithmEndpoints(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "catalog.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := algs.InitCatalog(db, filepath.Join("..", "..", "..", "build", "Alg-Trainers")); err != nil {
		t.Fatal(err)
	}
	service := &svc.Svc{DB: db, Cache: cache.New(cache.NoExpiration, 0)}
	router := gin.New()
	router.GET("/algorithm/", AlgorithmGroups(service))
	router.GET("/algorithm/:cubeID/:classID", AlgorithmGroupsWithCube(service))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/algorithm/", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("list status = %d", response.Code)
	}
	if got := response.Header().Get("Cache-Control"); got != "" {
		t.Fatalf("unexpected list Cache-Control %q", got)
	}
	var list AlgorithmGroupsResponse
	if err := json.Unmarshal(response.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(list.CubeKeys, "fto") || len(list.ClassMap["fto"]) != 9 {
		t.Fatalf("expected an FTO category with nine classes")
	}
	for _, class := range list.ClassMap["special"] {
		if slices.Contains([]string{"FTO L3T", "FTO TCP"}, class.Name) {
			t.Errorf("%s is still listed under special", class.Name)
		}
	}
	tests := []struct {
		name  string
		cube  string
		cases int
	}{
		{"FTO L3T", "fto", 179},
		{"FTO TCP", "fto", 18},
		{"FTO LL", "fto", 377},
		{"FTO LT", "fto", 151},
		{"FTO LBT", "fto", 95},
		{"FTO L3C", "fto", 5},
		{"FTO BTLT", "fto", 23},
		{"FTO FTLT", "fto", 23},
		{"FTO L6X", "fto", 35},
		{"LCE", "bigCube", 398},
		{"CLL", "222", 289},
		{"TEG1", "222", 497},
		{"TEG2", "222", 665},
		{"CBL", "222", 2183},
		{"EGLS1", "222", 1830},
		{"EGLS2", "222", 396},
		{"FakeBar", "222", 1306},
		{"F2L", "333", 41},
		{"ZBLS", "333", 305},
		{"1LLL", "333", 3397},
		{"CMLL", "333oh", 42},
		{"4x4 PLLP", "bigCube", 22},
		{"5x5 L2E", "bigCube", 13},
		{"6x6 L2C", "bigCube", 32},
		{"Pyraminx L4E", "special", 49},
		{"CS", "sq1", 170},
		{"CSP", "sq1", 179},
		{"CP", "sq1", 8},
		{"EP", "sq1", 100},
		{"OBL", "sq1", 185},
		{"PBL", "sq1", 967},
		{"PARITY", "sq1", 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := false
			for _, class := range list.ClassMap[tt.cube] {
				if class.Name == tt.name {
					found = true
					if class.CaseCount != tt.cases || class.Image != "" || class.Alg == "" {
						t.Errorf("invalid list entry for %s: count=%d", tt.name, class.CaseCount)
					}
				}
			}
			if !found {
				t.Fatalf("%s is missing from %s", tt.name, tt.cube)
			}
			response := httptest.NewRecorder()
			path := "/algorithm/" + tt.cube + "/" + url.PathEscape(tt.name)
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("detail status = %d", response.Code)
			}
			if got := response.Header().Get("Cache-Control"); got != "" {
				t.Fatalf("unexpected detail Cache-Control %q", got)
			}
			var detail AlgorithmGroupsWithCubeResponse
			if err := json.Unmarshal(response.Body.Bytes(), &detail); err != nil {
				t.Fatal(err)
			}
			if detail.Name != tt.name || countClassCases(detail.AlgorithmClass) != tt.cases {
				t.Errorf("unexpected detail name or case count for %s", tt.name)
			}
			if tt.cube == "sq1" && tt.name == "EP" {
				first := detail.Sets[0].AlgorithmGroups[0].Algorithms[0]
				if first.Name != "Solved / Solved" || len(first.Algs) != 0 {
					t.Error("the solved case should remain in the detail response with no moves")
				}
			}
		})
	}
	t.Run("solution rows retain IDs across reordering and detail cache hits", func(t *testing.T) {
		detail := func(path string) algs.AlgorithmClass {
			t.Helper()
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, path, nil))
			if resp.Code != http.StatusOK {
				t.Fatalf("GET %s: %s", path, resp.Body.String())
			}
			var class algs.AlgorithmClass
			if err := json.Unmarshal(resp.Body.Bytes(), &class); err != nil {
				t.Fatal(err)
			}
			return class
		}
		find := func(class algs.AlgorithmClass, name string, n int) algs.Algorithm {
			t.Helper()
			for _, set := range class.Sets {
				for _, group := range set.AlgorithmGroups {
					for _, c := range group.Algorithms {
						if c.Name == name && len(c.Algs) == n {
							if len(group.GroupID) != 16 || len(c.CaseID) != 16 || len(c.SolutionIDs) != n {
								t.Fatalf("%s has incomplete identifiers", name)
							}
							return c
						}
					}
				}
			}
			t.Fatalf("missing %s with %d slots", name, n)
			return algs.Algorithm{}
		}
		oh := detail("/algorithm/333oh/ZBLL")
		duplicate := find(oh, "ADN11", 3)
		if duplicate.Algs[1] == "" || duplicate.Algs[1] != duplicate.Algs[2] || duplicate.SolutionIDs[1] == nil || duplicate.SolutionIDs[2] == nil || *duplicate.SolutionIDs[1] == *duplicate.SolutionIDs[2] {
			t.Fatal("identical occurrences must have distinct solution IDs")
		}
		for _, id := range duplicate.SolutionIDs[1:] {
			if _, err := hex.DecodeString(*id); err != nil || len(*id) != 16 {
				t.Fatalf("invalid solution ID %q", *id)
			}
		}
		blank := find(oh, "2GLL11", 3)
		if blank.Algs[1] != "" || blank.SolutionIDs[1] != nil || blank.SolutionIDs[0] == nil || blank.SolutionIDs[2] == nil {
			t.Fatal("the empty slot must preserve alignment with a null public ID")
		}
		var blankRows int64
		if err := db.Table("formula_solutions").Where("case_id = ? AND position = 1 AND moves = ''", blank.CaseID).Count(&blankRows).Error; err != nil || blankRows != 1 {
			t.Fatalf("empty rows at position 1 = %d, err = %v", blankRows, err)
		}
		if c := find(detail("/algorithm/333/F2L"), "A+", 13); len(c.Scrambles) != 1 {
			t.Fatal("scrambles must remain case-level")
		}
		const oll = "/algorithm/333/OLL"
		before := find(detail(oll), "OLL 45", 6)
		if before.SolutionIDs[0] == nil {
			t.Fatal("missing ID for first OLL 45 solution")
		}
		if err := db.Exec("UPDATE formula_solutions SET position = 6 WHERE id = ?", *before.SolutionIDs[0]).Error; err != nil {
			t.Fatal(err)
		}
		if err := algs.InitCatalog(db, filepath.Join(t.TempDir(), "missing")); err != nil {
			t.Fatalf("restarted database unexpectedly needs trainer files: %v", err)
		}
		if cached := find(detail(oll), "OLL 45", 6); !reflect.DeepEqual(cached, before) {
			t.Fatal("expected the cached detail before invalidation")
		}
		service.Cache.Flush()
		edited := find(detail(oll), "OLL 45", 6)
		wantIDs := append(slices.Clone(before.SolutionIDs[1:]), before.SolutionIDs[0])
		wantMoves := append(slices.Clone(before.Algs[1:]), before.Algs[0])
		if edited.CaseID != before.CaseID || !reflect.DeepEqual(edited.SolutionIDs, wantIDs) || !slices.Equal(edited.Algs, wantMoves) {
			t.Fatal("changing position must reorder moves and IDs together without changing their values")
		}
	})

	t.Run("previews use the earliest playable case", func(t *testing.T) {
		class, err := algs.ReadClass(db, "333", "F2L")
		if err != nil {
			t.Fatal(err)
		}
		first := class.Sets[0].AlgorithmGroups[0].Algorithms[0]
		preview := func() string {
			t.Helper()
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/algorithm/", nil))
			if resp.Code != http.StatusOK {
				t.Fatalf("overview status = %d", resp.Code)
			}
			var list AlgorithmGroupsResponse
			if err := json.Unmarshal(resp.Body.Bytes(), &list); err != nil {
				t.Fatal(err)
			}
			for _, cl := range list.ClassMap["333"] {
				if cl.Name == "F2L" {
					return cl.Alg
				}
			}
			t.Fatal("missing F2L preview")
			return ""
		}
		cached := preview()
		if err := db.Exec("UPDATE formula_solutions SET moves = '' WHERE case_id = ?", first.CaseID).Error; err != nil {
			t.Fatal(err)
		}
		for _, tt := range []struct{ first, second, want string }{
			{" ", " ", "F R' F' R"},
			{"y R U", "z F U", "y R U"},
			{"y R U", "F U", "F U"},
		} {
			for i, moves := range []string{tt.first, tt.second} {
				if err := db.Exec("UPDATE formula_solutions SET moves = ? WHERE id = ?", moves, *first.SolutionIDs[i]).Error; err != nil {
					t.Fatal(err)
				}
			}
			if got := preview(); got != cached {
				t.Fatalf("cached preview = %q, want %q", got, cached)
			}
			service.Cache.Delete(catalogOverviewCache)
			cached = preview()
			if cached != tt.want {
				t.Fatalf("refreshed preview = %q, want %q", cached, tt.want)
			}
		}
	})
}

func countClassCases(class algs.AlgorithmClass) int {
	n := 0
	for _, set := range class.Sets {
		for _, group := range set.AlgorithmGroups {
			n += len(group.Algorithms)
		}
	}
	return n
}
