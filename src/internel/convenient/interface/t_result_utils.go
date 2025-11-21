package _interface

import (
	"sort"
)

// RankByValue 对 items 原地排序并设置竞赛排名（1,2,2,4,...）
func RankByValue[T any](
	items []T,
	getValue func(T) int, // 提取用于排名的整数值（如 Sor、Score 等）
	setRank func(*T, int), // 设置 Rank 字段
	descending bool, // true: 值越大排名越前（如分数）；false: 值越小排名越前（如用时）
) {
	if len(items) == 0 {
		return
	}

	// Step 1: 原地排序
	sort.Slice(items, func(i, j int) bool {
		vi := getValue(items[i])
		vj := getValue(items[j])
		if descending {
			return vi > vj // 降序：高分在前
		}
		return vi < vj // 升序：低值在前
	})

	// Step 2: 按组分配排名（标准竞赛排名：1,2,2,4,...）
	rank := 1
	i := 0
	n := len(items)

	for i < n {
		currentVal := getValue(items[i])
		j := i
		// 找出所有值等于 currentVal 的连续元素（因为已排序，相同值一定相邻）
		for j < n && getValue(items[j]) == currentVal {
			j++
		}
		// 给 [i, j) 范围内所有元素设置相同 rank
		for k := i; k < j; k++ {
			setRank(&items[k], rank)
		}
		// 下一个 rank = 已处理人数 + 1
		rank = j + 1
		i = j
	}
}
