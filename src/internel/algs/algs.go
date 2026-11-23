package algs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func loadStatic(basePath string) (map[string]CubeAlgorithms, error) {
	basePath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("resolve trainer directory: %w", err)
	}
	catalogPath := filepath.Join(filepath.Dir(basePath), "Algs", "catalog.json")
	loaded, err := readCatalogFiles(basePath, catalogPath)
	if err != nil {
		return nil, fmt.Errorf("load formulas (trainers: %s, catalog: %s): %w", basePath, catalogPath, err)
	}
	return loaded, nil
}

func readCatalogFiles(basePath, catalogPath string) (map[string]CubeAlgorithms, error) {
	if info, err := os.Stat(basePath); err != nil {
		return nil, fmt.Errorf("open trainer directory: %w", err)
	} else if !info.IsDir() {
		return nil, fmt.Errorf("trainer path is not a directory: %s", basePath)
	}
	libraryPath := filepath.Dir(catalogPath)
	var catalog []struct {
		Cube     string   `json:"cube"`
		Name     string   `json:"name"`
		Path     string   `json:"path"`
		Replaces []string `json:"replaces"`
		Cases    int      `json:"cases"`
	}
	if err := writeJsonToObj(catalogPath, &catalog); err != nil {
		return nil, fmt.Errorf("read formula catalog: %w", err)
	}
	if len(catalog) == 0 {
		return nil, fmt.Errorf("formula catalog is empty")
	}
	loaded := make(map[string]CubeAlgorithms)
	replaced := make(map[string]map[string]bool)
	for _, key := range CubeKeyList {
		loaded[key] = CubeAlgorithms{Cube: key, ClassKeys: []string{}, ClassList: []AlgorithmClass{}}
		replaced[key] = make(map[string]bool)
	}
	for _, entry := range catalog {
		cube, ok := loaded[entry.Cube]
		if !ok || entry.Name == "" || !filepath.IsLocal(entry.Path) {
			return nil, fmt.Errorf("invalid formula catalog entry: %q/%q", entry.Cube, entry.Name)
		}
		for _, name := range cube.ClassKeys {
			if name == entry.Name {
				return nil, fmt.Errorf("duplicate formula class: %s/%s", entry.Cube, entry.Name)
			}
		}
		data, err := ReadTrainerFiles(filepath.Join(libraryPath, entry.Path))
		if err != nil {
			return nil, fmt.Errorf("read %s/%s: %w", entry.Cube, entry.Name, err)
		}
		if len(data.AlgsInfo) != entry.Cases || len(data.SetKeys) == 0 {
			return nil, fmt.Errorf("incomplete formula set: %s/%s", entry.Cube, entry.Name)
		}
		cube.ClassKeys = append(cube.ClassKeys, entry.Name)
		cube.ClassList = append(cube.ClassList, *fileAlgToAlgorithmClass(entry.Name, data))
		loaded[entry.Cube] = cube
		replaced[entry.Cube][entry.Name] = true
		for _, name := range entry.Replaces {
			replaced[entry.Cube][name] = true
		}
	}
	for _, key := range CubeKeyList {
		cube := loaded[key]
		for _, trainer := range algsDataKey[key] {
			name := algsNameMap[trainer]
			if replaced[key][name] {
				continue
			}
			trainerPath := filepath.Join(basePath, trainer)
			// The legacy configuration also lists trainers that were never shipped.
			if _, err := os.Stat(trainerPath); os.IsNotExist(err) {
				continue
			}
			data, err := ReadTrainerFiles(trainerPath)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", trainer, err)
			}
			cube.ClassKeys = append(cube.ClassKeys, name)
			cube.ClassList = append(cube.ClassList, *fileAlgToAlgorithmClass(name, data))
		}
		loaded[key] = cube
	}
	return loaded, nil
}

func bestScrambles(in []string) []string {
	sort.Slice(in, func(i, j int) bool {
		return len(strings.Split(in[i], " ")) < len(strings.Split(in[j], ""))
	})
	return in
}

// 例如EG为一个class
func fileAlgToAlgorithmClass(name string, fileAlg *AlgorithmConfigWithTrainer) *AlgorithmClass {
	out := &AlgorithmClass{
		Name:    name,
		Sets:    make([]AlgorithmSet, 0),
		SetKeys: fileAlg.SetKeys,
	}

	// 大组 EG0, EG1, CLL, LEG
	for _, setKey := range fileAlg.SetKeys {

		set := AlgorithmSet{
			Name:            setKey,
			AlgorithmGroups: make([]AlgorithmGroup, 0),
			GroupsKeys:      fileAlg.AlgsetsInfo[setKey],
		}

		// 分组 EG1-H, EG1-U...
		groups := fileAlg.AlgsetsInfo[setKey]
		for _, groupKey := range groups {
			groupIndexs, ok := fileAlg.GroupsInfo[groupKey]
			if !ok {
				continue
			}

			group := AlgorithmGroup{
				Name:       groupKey,
				Algorithms: make([]Algorithm, 0),
			}

			// 具体公式
			for _, groupIndex := range groupIndexs {
				algKey := fmt.Sprintf("%d", groupIndex)

				alg, ok2 := fileAlg.AlgsInfo[algKey]
				if !ok2 {
					continue
				}
				algorithm := Algorithm{
					sourceKey: algKey,
					Name:      fmt.Sprintf("%+v", alg.Name),
					Algs:      alg.Algs,
					Scrambles: bestScrambles(fileAlg.Scrambles[algKey]),
				}
				group.Algorithms = append(group.Algorithms, algorithm)
			}
			set.AlgorithmGroups = append(set.AlgorithmGroups, group)
		}

		out.Sets = append(out.Sets, set)
	}

	return out
}
