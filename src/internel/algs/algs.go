package algs

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

var baseAlgs = make(map[string]CubeAlgorithms)

func Init(basePath string) error {
	baseAlgs = make(map[string]CubeAlgorithms)
	for key, subKeys := range algsDataKey {
		cube := builderCubeAlgorithms(basePath, subKeys)
		cube.Cube = key
		baseAlgs[key] = cube
	}
	return nil
}

func GetAlgorithms() map[string]CubeAlgorithms {
	return baseAlgs
}

func builderCubeAlgorithms(basePath string, keys []string) CubeAlgorithms {

	out := CubeAlgorithms{
		ClassKeys: make([]string, 0),
		ClassList: make([]AlgorithmClass, 0),
	}

	for _, key := range keys {
		filePaths := path.Join(basePath, key)
		fileAlgs, err := ReadTrainerFiles(filePaths)
		if err != nil {
			continue
		}

		out.ClassList = append(out.ClassList, *fileAlgToAlgorithmClass(key, fileAlgs))
		out.ClassKeys = append(out.ClassKeys, algsNameMap[key])
	}
	return out
}

func bestScrambles(in []string) []string {
	sort.Slice(in, func(i, j int) bool {
		return len(strings.Split(in[i], " ")) < len(strings.Split(in[j], ""))
	})
	return in
}

// 例如EG为一个class
func fileAlgToAlgorithmClass(fileKey string, fileAlg *AlgorithmConfigWithTrainer) *AlgorithmClass {
	out := &AlgorithmClass{
		Name:    algsNameMap[fileKey],
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
					Image:     fileAlg.Images[algKey],
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
