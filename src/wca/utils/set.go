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

func HasIntersection(a, b []string) bool {
	m := make(map[string]struct{})

	for _, v := range a {
		m[v] = struct{}{}
	}

	for _, v := range b {
		if _, ok := m[v]; ok {
			return true
		}
	}

	return false
}
