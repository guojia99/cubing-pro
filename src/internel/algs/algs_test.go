package algs

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestGetAlgorithms(t *testing.T) {
	data, err := loadStatic(filepath.Join("..", "..", "..", "build", "Alg-Trainers"))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(CubeKeyList, "fto") {
		t.Fatal("FTO is missing from the category list")
	}
	wantFTO := []string{"FTO L3T", "FTO TCP", "FTO LL", "FTO LT", "FTO LBT", "FTO L3C", "FTO BTLT", "FTO FTLT", "FTO L6X"}
	if !slices.Equal(data["fto"].ClassKeys, wantFTO) {
		t.Errorf("FTO classes = %v, want %v", data["fto"].ClassKeys, wantFTO)
	}
	if !slices.Equal(data["special"].ClassKeys, []string{"Pyraminx L4E"}) {
		t.Errorf("special classes = %v, want only Pyraminx L4E", data["special"].ClassKeys)
	}
	wantBig := []string{"4x4 OLLP", "4x4 PLLP", "5x5 L2E", "5x5 L2C", "6x6 L2E", "6x6 L2C", "LCE"}
	if !slices.Equal(data["bigCube"].ClassKeys, wantBig) {
		t.Errorf("bigCube classes = %v, want %v", data["bigCube"].ClassKeys, wantBig)
	}
	want222 := []string{"CLL", "EG1", "EG2", "LEG", "TCLL+", "TCLL-", "LS1", "LS2", "LS3", "LS4", "LS5", "LS6", "LS7", "LS8", "LS9", "TEG1", "FH", "TEG2", "CBL", "EGLS1", "EGLS2", "FakeBar", "OLL", "PBL"}
	if !slices.Equal(data["222"].ClassKeys, want222) {
		t.Errorf("222 classes = %v, want %v", data["222"].ClassKeys, want222)
	}
	wantSq1 := []string{"CS", "CSP", "CO", "EO", "CP", "EP", "OBL", "PBL", "PARITY"}
	if !slices.Equal(data["sq1"].ClassKeys, wantSq1) {
		t.Errorf("sq1 classes = %v, want %v", data["sq1"].ClassKeys, wantSq1)
	}
	if len(data["333"].ClassList) != 24 || !slices.Equal(data["333oh"].ClassKeys, []string{"CMLL", "OLL", "PLL", "ZBLL"}) {
		t.Error("expected 24 3x3 sets and OH CMLL alongside legacy OH sets")
	}
	for cube, value := range data {
		seen := make(map[string]bool)
		for _, class := range value.ClassList {
			if seen[class.Name] {
				t.Errorf("duplicate class %s/%s", cube, class.Name)
			}
			seen[class.Name] = true
			for _, set := range class.Sets {
				for _, group := range set.AlgorithmGroups {
					for _, algorithm := range group.Algorithms {
						if algorithm.Image != "" {
							t.Errorf("%s/%s still returns an image", cube, class.Name)
						}
					}
				}
			}
		}
	}
}

func TestInitRejectsBrokenLibrary(t *testing.T) {
	tests := []struct {
		name    string
		catalog string
	}{
		{"missing catalog", ""},
		{"invalid JSON", "["},
		{"empty catalog", "[]"},
		{"unknown category", `[{"cube":"unknown","name":"OLL","path":"algs/oll"}]`},
		{"path escape", `[{"cube":"333","name":"OLL","path":"../Alg-Trainers"}]`},
		{"missing set", `[{"cube":"333","name":"OLL","path":"algs/oll","cases":57}]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.Mkdir(filepath.Join(root, "Alg-Trainers"), 0755); err != nil {
				t.Fatal(err)
			}
			if tt.catalog != "" {
				if err := os.Mkdir(filepath.Join(root, "Algs"), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, "Algs", "catalog.json"), []byte(tt.catalog), 0644); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := loadStatic(filepath.Join(root, "Alg-Trainers")); err == nil {
				t.Fatal("broken library should fail initialization")
			}
		})
	}
}
