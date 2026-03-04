package wca

func buildEventOrderMap() map[string]int {
	orderMap := make(map[string]int)
	for i, event := range wcaEventsList {
		orderMap[event] = i
	}
	return orderMap
}
