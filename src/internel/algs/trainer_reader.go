package algs

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/tidwall/gjson"
)

// readJSONToOrderedStringMap 读取 JSON 文件，返回 map[string]string 和按原始顺序排列的 key 列表
// 要求 JSON 根是一个对象（{}）
func readJSONToOrderedStringMap(filePath string) (keys []string, err error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		err = fmt.Errorf("failed to read file: %w", err)
		return
	}

	// 确保是 JSON 对象
	if !gjson.ValidBytes(data) {
		err = fmt.Errorf("invalid JSON")
		return
	}

	result := gjson.ParseBytes(data)
	if !result.IsObject() {
		err = fmt.Errorf("JSON root is not an object")
		return
	}

	m := make(map[string]interface{})
	keys = make([]string, 0)

	// ForEach 保证按 JSON 中字段的原始顺序遍历
	result.ForEach(func(key, value gjson.Result) bool {
		k := key.String()
		m[k] = nil
		keys = append(keys, k)
		return true // 继续遍历
	})

	return
}

func writeJsonToObj(filePath string, obj interface{}) error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, obj)
}

func ReadTrainerFiles(filePath string) (*AlgorithmConfigWithTrainer, error) {
	var out = &AlgorithmConfigWithTrainer{
		Name:        filepath.Base(filePath),
		AlgsInfo:    make(map[string]TrainerAlgorithm),
		AlgsetsInfo: make(map[string][]string),
		GroupsInfo:  make(map[string][]int),
		Images:      make(map[string]string),
		Scrambles:   make(map[string][]string),
	}

	if err := writeJsonToObj(path.Join(filePath, algsInfoFile), &out.AlgsInfo); err != nil {
		return nil, err
	}
	if err := writeJsonToObj(path.Join(filePath, algsetsInfoFile), &out.AlgsetsInfo); err != nil {
		return nil, err
	}
	if err := writeJsonToObj(path.Join(filePath, algImagesInfoFile), &out.Images); err != nil {
		return nil, err
	}
	if err := writeJsonToObj(path.Join(filePath, algGroupFile), &out.GroupsInfo); err != nil {
		return nil, err
	}
	if err := writeJsonToObj(path.Join(filePath, algScramblesInfoFile), &out.Scrambles); err != nil {
		return nil, err
	}

	keys, err := readJSONToOrderedStringMap(path.Join(filePath, algsetsInfoFile))
	if err != nil {
		return nil, err
	}
	out.SetKeys = keys

	return out, nil
}
