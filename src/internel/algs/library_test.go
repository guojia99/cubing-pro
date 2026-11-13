package algs

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestImportedLibrary(t *testing.T) {
	root := filepath.Join("..", "..", "..", "build")
	if err := Init(filepath.Join(root, "Alg-Trainers")); err != nil {
		t.Fatal(err)
	}
	var catalog []struct {
		Cube   string            `json:"cube"`
		Name   string            `json:"name"`
		Path   string            `json:"path"`
		Cases  int               `json:"cases"`
		Hashes map[string]string `json:"sha256"`
	}
	if err := writeJsonToObj(filepath.Join(root, "Algs", "catalog.json"), &catalog); err != nil {
		t.Fatal(err)
	}
	if len(catalog) != 63 {
		t.Fatalf("catalog has %d sets, want 63", len(catalog))
	}
	total := 0
	for _, entry := range catalog {
		total += entry.Cases
		t.Run(entry.Cube+"/"+entry.Name, func(t *testing.T) {
			path := filepath.Join(root, "Algs", entry.Path)
			if len(entry.Hashes) != 4 {
				t.Fatal("expected hashes for four data files")
			}
			for name, want := range entry.Hashes {
				data, err := os.ReadFile(filepath.Join(path, name))
				if err != nil {
					t.Fatal(err)
				}
				if fmt.Sprintf("%x", sha256.Sum256(data)) != want {
					t.Errorf("%s differs from the imported snapshot", name)
				}
			}
			if _, err := os.Stat(filepath.Join(path, "combined.json")); !os.IsNotExist(err) {
				t.Fatal("image payload should not be imported")
			}
			data, err := ReadTrainerFiles(path)
			if err != nil {
				t.Fatal(err)
			}
			seen := make(map[string]bool)
			for _, set := range data.SetKeys {
				for _, group := range data.AlgsetsInfo[set] {
					if len(data.GroupsInfo[group]) == 0 {
						t.Errorf("group %s is empty or missing", group)
					}
					for _, id := range data.GroupsInfo[group] {
						key := fmt.Sprint(id)
						alg, ok := data.AlgsInfo[key]
						if !ok || alg.Algset != set || alg.Group != group || alg.Name == nil {
							t.Errorf("case %s has inconsistent metadata", key)
						}
						if seen[key] {
							t.Errorf("case %s is referenced more than once", key)
						}
						seen[key] = true
						if _, ok := data.Scrambles[key]; !ok {
							t.Errorf("case %s has no scramble entry", key)
						}
					}
				}
			}
			if len(seen) != entry.Cases || len(data.AlgsInfo) != entry.Cases {
				t.Errorf("case count mismatch: referenced=%d loaded=%d want=%d", len(seen), len(data.AlgsInfo), entry.Cases)
			}
			want := fileAlgToAlgorithmClass(entry.Name, data)
			var actual *AlgorithmClass
			cube := GetAlgorithms()[entry.Cube]
			for i := range cube.ClassList {
				if cube.ClassList[i].Name == entry.Name {
					actual = &cube.ClassList[i]
				}
			}
			// Compare the serialized API contract, including solved cases and formula order.
			gotJSON, _ := json.Marshal(actual)
			wantJSON, _ := json.Marshal(want)
			if !reflect.DeepEqual(gotJSON, wantJSON) {
				t.Error("published class does not match the selected new-library source")
			}
		})
	}
	if total != 17172 {
		t.Errorf("imported cases = %d, want 17172", total)
	}
}
