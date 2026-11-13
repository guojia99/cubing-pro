package algs

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

var baseAlgs = make(map[string]CubeAlgorithms)

func Init(basePath string) error {
	libraryPath := filepath.Join(filepath.Dir(filepath.Clean(basePath)), "Algs")
	var catalog []struct {
		Cube     string   `json:"cube"`
		Name     string   `json:"name"`
		Path     string   `json:"path"`
		Replaces []string `json:"replaces"`
		Cases    int      `json:"cases"`
	}
	if err := writeJsonToObj(filepath.Join(libraryPath, "catalog.json"), &catalog); err != nil {
		return fmt.Errorf("read formula catalog: %w", err)
	}
	if len(catalog) == 0 {
		return fmt.Errorf("formula catalog is empty")
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
			return fmt.Errorf("invalid formula catalog entry: %q/%q", entry.Cube, entry.Name)
		}
		for _, name := range cube.ClassKeys {
			if name == entry.Name {
				return fmt.Errorf("duplicate formula class: %s/%s", entry.Cube, entry.Name)
			}
		}
		data, err := ReadTrainerFiles(filepath.Join(libraryPath, entry.Path))
		if err != nil {
			return fmt.Errorf("read %s/%s: %w", entry.Cube, entry.Name, err)
		}
		if len(data.AlgsInfo) != entry.Cases || len(data.SetKeys) == 0 {
			return fmt.Errorf("incomplete formula set: %s/%s", entry.Cube, entry.Name)
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
			data, err := ReadTrainerFiles(filepath.Join(basePath, trainer))
			if err != nil {
				continue
			}
			cube.ClassKeys = append(cube.ClassKeys, name)
			cube.ClassList = append(cube.ClassList, *fileAlgToAlgorithmClass(name, data))
		}
		loaded[key] = cube
	}
	baseAlgs = loaded
	return nil
}

func GetAlgorithms() map[string]CubeAlgorithms {
	return baseAlgs
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
