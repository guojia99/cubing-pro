package utils_tool

func IntersectAll(lists [][]string) []string {
	if len(lists) == 0 {
		return nil
	}

	count := make(map[string]int)

	for _, list := range lists {
		seen := make(map[string]struct{}) // 防止单个 list 内重复
		for _, id := range list {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			count[id]++
		}
	}

	var result []string
	for id, c := range count {
		if c == len(lists) {
			result = append(result, id)
		}
	}

	return result
}
