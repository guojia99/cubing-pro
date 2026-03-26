package wca

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
)

func buildEventOrderMap() map[string]int {
	orderMap := make(map[string]int)
	for i, event := range wcaEventsList {
		orderMap[event] = i
	}
	return orderMap
}

func buildIndexMap(events []string) map[string]uint64 {
	m := make(map[string]uint64)
	for i, e := range events {
		m[e] = uint64(i)
	}
	return m
}
func printJson(i interface{}) {
	data, err := jsoniter.MarshalIndent(&i, "", "  ")
	fmt.Println(err)
	fmt.Printf("%+v\n", i)
	fmt.Println(string(data))
}

//	for comb := range combinationsStream(wcaEventsList, 2, 16) {
//		// 按需处理（写库 / 过滤 / 计算）
//	}
func combinationsStream(arr []string, minLen, maxLen int) <-chan []string {
	ch := make(chan []string)

	go func() {
		defer close(ch)

		var path []string

		var dfs func(start int)
		dfs = func(start int) {
			if len(path) >= minLen && len(path) <= maxLen {
				comb := make([]string, len(path))
				copy(comb, path)
				ch <- comb
			}

			if len(path) >= maxLen {
				return
			}

			for i := start; i < len(arr); i++ {
				path = append(path, arr[i])
				dfs(i + 1)
				path = path[:len(path)-1]
			}
		}

		dfs(0)
	}()
	return ch
}

func encodeEvents(events []string) uint64 {
	var mask uint64 = 0
	for _, e := range events {
		if idx, ok := wcaEventMap[e]; ok {
			mask |= 1 << idx
		}
	}
	return mask
}

func decodeEvents(mask uint64) []string {
	var res []string
	for i := 0; i < len(wcaEventsList); i++ {
		if mask&(1<<i) != 0 {
			res = append(res, wcaEventsList[i])
		}
	}
	return res
}

func checkHasEvent(allEventCode uint64, personCode uint64) bool {
	if personCode == 0 {
		return false
	}
	return allEventCode&personCode != 0
}
