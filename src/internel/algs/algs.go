package algs

import (
	"fmt"
	"path"
)

var baseAlgs = make(map[string]CubeAlgorithms)
var CubeKeyList = []string{
	"222",
	"333",
	"333oh",
	"minx",
	"sq1",
	"skewb",
}
var algsDataKey = map[string][]string{
	"222": {
		"2x2-EG-Trainer",
		"2x2-TCLL-Trainer",
		"2x2-TEG-Trainer",
		"2x2-FH-Trainer",
		"2x2-LS-Trainer",
	},
	"333": {
		"3x3-OLL-Trainer",
		"3x3-PLL-Trainer",
		"3x3-CMLL-Trainer",
		"3x3-ZBLL-Trainer",
		"3x3-ZBLS-Trainer",
	},
	"333oh": {
		"3x3-OH-CMLL-Trainer",
		"3x3-OH-OLL-Trainer",
		"3x3-OH-PLL-Trainer",
		"3x3-OH-ZBLL-Trainer",
	},
	"minx": {
		"Megaminx-OLL-Trainer",
		"Megaminx-PLL-Trainer",
	},
	"sq1": {
		"Sq1-CPEP-Trainer",
		"Sq1-OBL-Trainer",
		"Sq1-PBL-Trainer",
	},
	"skewb": {
		"Skewb-NS2-Trainer",
	},
}

var algsNameMap = map[string]string{
	"2x2-EG-Trainer":   "EG",
	"2x2-FH-Trainer":   "FH",
	"2x2-LS-Trainer":   "LS",
	"2x2-TCLL-Trainer": "TCLL",
	"2x2-TEG-Trainer":  "TEG",

	"3x3-OLL-Trainer":  "OLL",
	"3x3-PLL-Trainer":  "PLL",
	"3x3-CMLL-Trainer": "CMLL",
	"3x3-ZBLL-Trainer": "ZBLL",
	"3x3-ZBLS-Trainer": "ZBLS",

	"3x3-OH-CMLL-Trainer": "CMLL",
	"3x3-OH-OLL-Trainer":  "OLL",
	"3x3-OH-PLL-Trainer":  "PLL",
	"3x3-OH-ZBLL-Trainer": "ZBLL",

	"Megaminx-OLL-Trainer": "OLL",
	"Megaminx-PLL-Trainer": "PLL",

	"Sq1-CPEP-Trainer": "CPEP",
	"Sq1-OBL-Trainer":  "OBL",
	"Sq1-PBL-Trainer":  "PBL",

	"Skewb-NS2-Trainer": "NS",
}

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
					Scrambles: fileAlg.Scrambles[algKey],
				}
				group.Algorithms = append(group.Algorithms, algorithm)
			}
			set.AlgorithmGroups = append(set.AlgorithmGroups, group)
		}

		out.Sets = append(out.Sets, set)
	}

	return out
}
