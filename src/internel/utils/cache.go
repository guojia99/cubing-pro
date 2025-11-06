package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

// MakeCacheKey 生成一个基于输入参数的唯一缓存 key
// 输入应尽量是确定性的（例如 slice 已排序，map 转为 sorted slice of pairs）
func MakeCacheKey(args ...interface{}) (string, error) {
	// 规范化处理：对顶层 slice 类型尝试排序（仅限 []string 和 []int 等简单类型）
	normalizedArgs := make([]interface{}, len(args))
	for i, arg := range args {
		switch v := arg.(type) {
		case []string:
			cp := make([]string, len(v))
			copy(cp, v)
			sort.Strings(cp)
			normalizedArgs[i] = cp
		case []int:
			cp := make([]int, len(v))
			copy(cp, v)
			sort.Ints(cp)
			normalizedArgs[i] = cp
		// 可根据需要扩展其他可排序类型，如 []int64, 自定义 ID 列表等
		default:
			// 其他类型（包括 struct, int, string 等）直接使用
			// 注意：map[string]interface{} 在 JSON 中顺序不确定，不建议直接传入
			normalizedArgs[i] = arg
		}
	}

	data, err := json.Marshal(normalizedArgs)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}
