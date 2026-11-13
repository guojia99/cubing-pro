package algs

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestReadTrainerFiles(t *testing.T) {
	data, err := ReadTrainerFiles(filepath.Join("..", "..", "..", "build", "Alg-Trainers", "2x2-EG-Trainer"))
	if err != nil {
		t.Fatal(err)
	}

	if len(data.SetKeys) == 0 || len(data.AlgsInfo) == 0 {
		t.Fatal("EG trainer loaded without sets or cases")
	}
}

func TestSyncedTrainerFiles(t *testing.T) {
	tests := []struct {
		name  string
		cases int
	}{
		{"FTO-LL-Trainer", 377},
		{"FTO-LT-Trainer", 151},
		{"FTO-LBT-Trainer", 95},
		{"FTO-L3C-Trainer", 5},
		{"FTO-BTLT-Trainer", 23},
		{"FTO-FTLT-Trainer", 23},
		{"FTO-L6X-Trainer", 35},
		{"Octaminx-L3T-Trainer", 179},
		{"Octaminx-TCP-Trainer", 18},
		{"5x5-LCE-Trainer", 398},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ReadTrainerFiles(filepath.Join("..", "..", "..", "build", "Alg-Trainers", tt.name))
			if err != nil {
				t.Fatal(err)
			}
			if len(data.AlgsInfo) != tt.cases {
				t.Fatalf("cases = %d, want %d", len(data.AlgsInfo), tt.cases)
			}
			seen := make(map[string]bool)
			for _, set := range data.SetKeys {
				groups := data.AlgsetsInfo[set]
				if len(groups) == 0 {
					t.Fatalf("set %q has no groups", set)
				}
				for _, group := range groups {
					ids := data.GroupsInfo[group]
					if len(ids) == 0 {
						t.Fatalf("group %q has no cases", group)
					}
					for _, id := range ids {
						key := fmt.Sprint(id)
						alg, ok := data.AlgsInfo[key]
						if !ok || len(alg.Algs) == 0 || alg.Name == nil || set+" "+alg.Group != group || alg.Algset != set {
							t.Errorf("case %s has missing or inconsistent metadata", key)
						}
						if len(data.Scrambles[key]) == 0 {
							t.Errorf("case %s is missing its scrambles", key)
						}
						for _, formula := range alg.Algs {
							if formula == "" {
								t.Errorf("case %s has an empty formula", key)
							}
						}
						if seen[key] {
							t.Errorf("case %s is referenced more than once", key)
						}
						seen[key] = true
					}
				}
			}
			if len(seen) != tt.cases {
				t.Errorf("referenced cases = %d, want %d", len(seen), tt.cases)
			}
			if len(data.Images) != 0 {
				t.Error("trainer images should not be loaded")
			}
		})
	}
}
