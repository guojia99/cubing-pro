package event

import "sort"

var routes = func() []RouteType {
	var out []RouteType
	for key := range routeMaps {
		out = append(out, key)
	}

	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}()

func Routes() []RouteType {
	return routes
}
