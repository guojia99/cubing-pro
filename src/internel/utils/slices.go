package utils

func RemoveRepeatedElement[S ~[]E, E comparable](s S) S {
	result := make([]E, 0)
	m := make(map[E]bool)
	for _, v := range s {
		if _, ok := m[v]; !ok {
			result = append(result, v)
			m[v] = true
		}
	}
	return result
}
