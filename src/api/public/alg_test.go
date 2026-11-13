package public

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"slices"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/algs"
)

func TestSyncedAlgorithmEndpoints(t *testing.T) {
	if err := algs.Init(filepath.Join("..", "..", "..", "build", "Alg-Trainers")); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.GET("/algorithm/", AlgorithmGroups(nil))
	router.GET("/algorithm/:cubeID/:classID", AlgorithmGroupsWithCube(nil))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/algorithm/", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("list status = %d", response.Code)
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
}

func TestPreviewAlgorithm(t *testing.T) {
	tests := []struct {
		name     string
		formulas [][]string
		want     string
	}{
		{"empty class", nil, ""},
		{"solved case only", [][]string{nil}, ""},
		{"skip solved and blank formulas", [][]string{nil, {" ", "R U R'"}}, "R U R'"},
		{"prefer no rotations", [][]string{{"y R U R'", "R U R'"}}, "R U R'"},
		{"rotation fallback", [][]string{{"y R U R'"}}, "y R U R'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			class := algs.AlgorithmClass{}
			if tt.formulas != nil {
				group := algs.AlgorithmGroup{}
				for _, formulas := range tt.formulas {
					group.Algorithms = append(group.Algorithms, algs.Algorithm{Algs: formulas})
				}
				class.Sets = []algs.AlgorithmSet{{AlgorithmGroups: []algs.AlgorithmGroup{group}}}
			}
			if got := previewAlgorithm(class); got != tt.want {
				t.Errorf("preview = %q, want %q", got, tt.want)
			}
		})
	}
}
